// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"fmt"
	"time"

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
	update, err := u.checkForUpdate(ctx, options)
	if err != nil {
		return nil, err
	}
	if update == nil {
		// No update available
		return nil, nil
	}

	// Try to download (cache) asset before prompt so the update happens immediately
	// after the user chooses apply. In the case of linux we won't have an asset.
	if err := u.downloadAsset(update.Asset, options); err != nil {
		u.log.Warningf("Unable to download asset: %s", err)
	}

	// Prompt for update
	updateAction, err := u.promptForUpdateAction(ctx, *update, options)
	if err != nil {
		return update, err
	}
	switch updateAction {
	case UpdateActionApply, UpdateActionAuto:
		// Continue
	case UpdateActionSnooze:
		return update, cancelError("Snoozed update")
	case UpdateActionCancel:
		return update, cancelError("Canceled by user")
	case UpdateActionError:
		return update, fmt.Errorf("Error in update prompt: %s", err)
	}

	// Linux updates don't have assets so it's ok to prompt for update above before
	// we check for nil asset
	if update.Asset == nil || update.Asset.URL == "" {
		u.log.Info("No update asset to apply")
		return update, nil
	}
	if update.Asset.LocalPath == "" {
		return update, fmt.Errorf("No local asset path")
	}

	// Ensure asset still exists in case it got pulled while we were waiting in
	// the prompt.
	exists, err := util.URLExists(update.Asset.URL, time.Minute, u.log)
	if err != nil {
		return update, err
	}
	if !exists {
		return update, fmt.Errorf("Asset no longer exists: %s", update.Asset.URL)
	}

	u.log.Infof("Verify asset: %s", update.Asset.LocalPath)
	if err := ctx.Verify(*update); err != nil {
		return update, err
	}

	if err := ctx.BeforeApply(*update); err != nil {
		return update, err
	}

	if err := u.platformApplyUpdate(*update, options); err != nil {
		return update, err
	}

	if err := ctx.AfterApply(*update); err != nil {
		return update, err
	}

	if err := ctx.Restart(); err != nil {
		return update, err
	}

	return update, nil
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
	}

	return updatePromptResponse.Action, nil
}
