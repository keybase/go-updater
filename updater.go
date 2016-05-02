// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"fmt"

	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater/util"
)

// Version is the updater version
const Version = "0.2.1"

// Updater knows how to find and apply updates
type Updater struct {
	source UpdateSource
	config Config
	log    logging.Logger
}

// UpdateSource defines where the updater can find updates
type UpdateSource interface {
	// Description is a short description about the update source
	Description() string
	// FindUpdate finds an update given options
	FindUpdate(options UpdateOptions) (*Update, error)
}

// Context defines state during an update session
type Context interface {
	GetUpdateUI() (UpdateUI, error)
	UpdateOptions() UpdateOptions
	Verify(update Update) error
	BeforeApply(update Update) error
	AfterApply(update Update) error
	Restart() error
	ReportError(err error, options UpdateOptions)
	ReportAction(action UpdateAction, options UpdateOptions)
}

// Config defines configuration for the Updater
type Config interface {
	GetUpdateAuto() (bool, bool)
	SetUpdateAuto(b bool) error
	GetInstallID() string
	SetInstallID(installID string) error
}

// NewUpdater constructs an Updater
func NewUpdater(source UpdateSource, config Config, log logging.Logger) *Updater {
	return &Updater{
		source: source,
		config: config,
		log:    log,
	}
}

// Update checks, downloads and performs an update
func (u *Updater) Update(ctx Context) (*Update, error) {
	options := ctx.UpdateOptions()
	update, err := u.update(ctx, options)
	if err != nil {
		ctx.ReportError(err, options)
	}
	return update, err
}

func (u *Updater) update(ctx Context, options UpdateOptions) (*Update, error) {
	update, err := u.checkForUpdate(ctx, options)
	if err != nil {
		return nil, findErr(err)
	}
	if update == nil {
		// No update available
		return nil, nil
	}

	// Prompt for update
	updateAction, err := u.promptForUpdateAction(ctx, *update, options)
	if err != nil {
		return update, promptErr(err)
	}
	switch updateAction {
	case UpdateActionApply:
		ctx.ReportAction(UpdateActionApply, options)
	case UpdateActionAuto:
		ctx.ReportAction(UpdateActionAuto, options)
	case UpdateActionSnooze:
		ctx.ReportAction(UpdateActionSnooze, options)
		return update, cancelErr(fmt.Errorf("Snoozed update"))
	case UpdateActionCancel:
		ctx.ReportAction(UpdateActionCancel, options)
		return update, cancelErr(fmt.Errorf("Canceled"))
	case UpdateActionError:
		return update, promptErr(fmt.Errorf("Unknown prompt error"))
	}

	// Linux updates don't have assets so it's ok to prompt for update above before
	// we check for nil asset
	if update.Asset == nil || update.Asset.URL == "" {
		u.log.Info("No update asset to apply")
		return update, nil
	}

	if err := u.downloadAsset(update.Asset, options); err != nil {
		return update, downloadErr(err)
	}

	u.log.Infof("Verify asset: %s", update.Asset.LocalPath)
	if err := ctx.Verify(*update); err != nil {
		return update, verifyErr(err)
	}

	if err := u.apply(ctx, *update, options); err != nil {
		return update, err
	}

	if err := ctx.Restart(); err != nil {
		return update, restartErr(err)
	}

	return update, nil
}

func (u *Updater) apply(ctx Context, update Update, options UpdateOptions) error {
	if err := ctx.BeforeApply(update); err != nil {
		return applyErr(err)
	}

	if err := ctx.BeforeApply(update); err != nil {
		return applyErr(err)
	}

	if err := u.platformApplyUpdate(update, options); err != nil {
		return applyErr(err)
	}

	if err := ctx.AfterApply(update); err != nil {
		return applyErr(err)
	}

	return nil
}

// downloadAsset will download the update to a temporary path (if not cached),
// check the digest, and set the LocalPath property on the asset.
func (u *Updater) downloadAsset(asset *Asset, options UpdateOptions) error {
	if asset == nil {
		return fmt.Errorf("No asset to download")
	}
	downloadOptions := util.DownloadURLOptions{
		Digest:        asset.Digest,
		RequireDigest: true,
		UseETag:       true,
		Log:           u.log,
	}

	downloadPath := util.TempPath("", asset.Name+".")
	if err := util.DownloadURL(asset.URL, downloadPath, downloadOptions); err != nil {
		return err
	}

	asset.LocalPath = downloadPath
	return nil
}

// checkForUpdate checks a update source (like a remote API) for an update.
// It may set an InstallID, if the server tells us to.
func (u *Updater) checkForUpdate(ctx Context, options UpdateOptions) (*Update, error) {
	u.log.Infof("Checking for update, current version is %s", options.Version)
	u.log.Infof("Using updater source: %s", u.source.Description())
	u.log.Debugf("Using options: %#v", options)

	update, findErr := u.source.FindUpdate(options)
	if findErr != nil {
		return nil, findErr
	}
	if update == nil {
		return nil, nil
	}

	// Save InstallID if we received one
	if update.InstallID != "" {
		if err := u.config.SetInstallID(update.InstallID); err != nil {
			u.log.Warningf("Error saving install ID: %s", err)
			ctx.ReportError(configErr(fmt.Errorf("Error saving install ID: %s", err)), options)
		}
	}

	u.log.Infof("Got update with version: %s", update.Version)
	return update, nil
}

// promptForUpdateAction prompts the user for permission to apply an update
func (u *Updater) promptForUpdateAction(ctx Context, update Update, options UpdateOptions) (UpdateAction, error) {
	u.log.Debug("Prompt for update")

	auto, autoSet := u.config.GetUpdateAuto()
	if auto {
		u.log.Debug("Auto updates enabled")
		return UpdateActionAuto, nil
	}

	updateUI, err := ctx.GetUpdateUI()
	if err != nil {
		return UpdateActionError, err
	}

	// If auto update never set, default to true
	autoUpdate := !autoSet
	promptOptions := UpdatePromptOptions{AutoUpdate: autoUpdate}
	updatePromptResponse, err := updateUI.UpdatePrompt(update, options, promptOptions)
	if err != nil {
		return UpdateActionError, err
	}

	u.log.Debugf("Update prompt response: %#v", updatePromptResponse)
	if err := u.config.SetUpdateAuto(updatePromptResponse.AutoUpdate); err != nil {
		u.log.Warningf("Error setting auto preference: %s", err)
		ctx.ReportError(configErr(fmt.Errorf("Error setting auto preference: %s", err)), options)
	}

	return updatePromptResponse.Action, nil
}
