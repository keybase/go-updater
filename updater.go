// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/keybase/client/go/libkb"
	"github.com/keybase/client/go/logger"
	zip "github.com/keybase/client/go/tools/zip"
	keybase1 "github.com/keybase/go-updater/protocol"
	"github.com/keybase/go-updater/sources"
	"golang.org/x/net/context"
)

// Version is the updater version
const Version = "0.2.1"

// Updater knows how to find and apply updates
type Updater struct {
	source       sources.UpdateSource
	config       Config
	log          logger.Logger
	callGroup    Group
	cancelPrompt context.CancelFunc
}

// UpdateUI defines the interface to UI components
type UpdateUI interface {
	keybase1.UpdateUI
}

// Context defines state during an update session
type Context interface {
	GetUpdateUI() (UpdateUI, error)
	UpdateOptions() (keybase1.UpdateOptions, error)
	Verify(r io.Reader, signature string) error
}

// Config defines configuration for the Updater
type Config interface {
	GetUpdatePreferenceAuto() (bool, bool)
	SetUpdatePreferenceAuto(b bool) error
}

// CanceledError is for when an update is canceled
type CanceledError struct {
	message string
}

// Error returns canceled message
func (c CanceledError) Error() string {
	return c.message
}

// NewCanceledError constructs a CanceledError
func NewCanceledError(message string) CanceledError {
	return CanceledError{message}
}

// NewUpdater constructs an Updater
func NewUpdater(source sources.UpdateSource, config Config, log logger.Logger) *Updater {
	return &Updater{
		source: source,
		config: config,
		log:    log,
	}
}

func (u *Updater) checkForUpdate(options keybase1.UpdateOptions, skipAssetDownload bool) (update *keybase1.Update, err error) {
	u.log.Info("Checking for update, current version is %s", options.Version)

	u.log.Info("Using updater source: %s", u.source.Description())
	u.log.Debug("Using options: %#v", options)
	update, err = u.source.FindUpdate(options)
	if err != nil || update == nil {
		return
	}

	u.log.Info("Checking update with version: %s", update.Version)
	updateSemVersion, err := semver.Make(update.Version)
	if err != nil {
		return
	}

	currentSemVersion, err := semver.Make(options.Version)
	if err != nil {
		return
	}

	if updateSemVersion.EQ(currentSemVersion) {
		// Versions are the same, we are up to date
		u.log.Info("Update matches current version: %s = %s", updateSemVersion, currentSemVersion)
		if !options.Force {
			update = nil
			return
		}
	} else if updateSemVersion.LT(currentSemVersion) {
		u.log.Info("Update is older version: %s < %s", updateSemVersion, currentSemVersion)
		if !options.Force {
			update = nil
			return
		}
	}

	if !skipAssetDownload && update.Asset != nil {
		downloadPath, _, dlerr := u.downloadAsset(*update.Asset)
		if dlerr != nil {
			err = dlerr
			return
		}
		update.Asset.LocalPath = downloadPath
	}

	return
}

func computeEtag(path string) (string, error) {
	var result []byte
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(result)), nil
}

func (u *Updater) checkDigest(digest string, localPath string) error {
	if digest == "" {
		return fmt.Errorf("Missing digest")
	}
	calcDigest, err := libkb.DigestForFileAtPath(localPath)
	if err != nil {
		return err
	}
	if calcDigest != digest {
		return fmt.Errorf("Invalid digest: %s != %s (%s)", calcDigest, digest, localPath)
	}
	u.log.Info("Verified digest: %s (%s)", digest, localPath)
	return nil
}

func (u *Updater) downloadAsset(asset keybase1.Asset) (fpath string, cached bool, err error) {
	url, err := url.Parse(asset.URL)
	if err != nil {
		return
	}

	filename := asset.Name
	fpath = pathForUpdaterFilename(filename)
	err = libkb.MakeParentDirs(fpath)
	if err != nil {
		return
	}

	if url.Scheme == "file" {
		// This is only used for testing, where "file://" is hardcoded.
		// "file:\\" doesn't work on Windows here.
		localpath := asset.URL[7:]

		err = copyFile(localpath, fpath)
		if err != nil {
			return
		}

		if derr := u.checkDigest(asset.Digest, fpath); derr != nil {
			err = derr
			return
		}

		u.log.Info("Using local path: %s", fpath)
		return
	}

	etag := ""
	if _, err = os.Stat(fpath); err == nil {
		etag, err = computeEtag(fpath)
		if err != nil {
			return
		}
	}

	req, _ := http.NewRequest("GET", url.String(), nil)
	if etag != "" {
		u.log.Info("Using etag: %s", etag)
		req.Header.Set("If-None-Match", etag)
	}
	timeout := 20 * time.Minute
	client := &http.Client{
		Timeout: timeout,
	}
	u.log.Info("Request %s", url.String())
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	if resp == nil {
		err = fmt.Errorf("No response")
		return
	}
	defer func() { _ = libkb.DiscardAndCloseBody(resp) }()
	if resp.StatusCode == http.StatusNotModified {
		u.log.Info("Using cached file: %s", fpath)
		cached = true
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Responded with %s", resp.Status)
		return
	}

	savePath := fmt.Sprintf("%s.download", fpath)
	if _, ferr := os.Stat(savePath); ferr == nil {
		u.log.Info("Removing existing partial download: %s", savePath)
		err = os.Remove(savePath)
		if err != nil {
			return
		}
	}

	err = u.save(savePath, *resp)
	if err != nil {
		return
	}

	if derr := u.checkDigest(asset.Digest, savePath); derr != nil {
		err = derr
		return
	}

	if _, err = os.Stat(fpath); err == nil {
		u.log.Info("Removing existing download: %s", fpath)
		err = os.Remove(fpath)
		if err != nil {
			return
		}
	}

	u.log.Info("Moving %s to %s", filepath.Base(savePath), filepath.Base(fpath))

	err = os.Rename(savePath, fpath)
	return
}

func (u *Updater) save(savePath string, resp http.Response) error {
	file, err := os.OpenFile(savePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, libkb.PermFile)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	u.log.Info("Downloading to %s", savePath)
	n, err := io.Copy(file, resp.Body)
	if err == nil {
		u.log.Info("Downloaded %d bytes", n)
	}
	return err
}

func (u *Updater) unpack(filename string) (string, error) {
	u.log.Debug("Unpack %s", filename)
	if !strings.HasSuffix(filename, ".zip") {
		u.log.Debug("File isn't compressed, so won't unzip: %q", filename)
		return filename, nil
	}

	unzipDestination := unzipDestination(filename)
	if _, ferr := os.Stat(unzipDestination); ferr == nil {
		u.log.Info("Removing existing unzip destination: %s", unzipDestination)
		err := os.RemoveAll(unzipDestination)
		if err != nil {
			return "", nil
		}
	}

	u.log.Info("Unzipping %q -> %q", filename, unzipDestination)
	err := zip.Unzip(filename, unzipDestination)
	if err != nil {
		u.log.Errorf("Don't know how to unpack: %s", filename)
		return "", err
	}

	return unzipDestination, nil
}

func (u *Updater) ensureAssetExists(asset keybase1.Asset) error {
	if !strings.HasPrefix(asset.URL, "http://") && !strings.HasPrefix(asset.URL, "https://") {
		u.log.Debug("Skipping re-check for non-http asset")
		return nil
	}
	u.log.Debug("Checking asset still exists") // In case the update was revoked
	resp, err := http.Head(asset.URL)
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("No response")
	}
	defer func() { _ = libkb.DiscardAndCloseBody(resp) }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Update is no longer available (%d)", resp.StatusCode)
	}
	return nil
}

func (u *Updater) checkUpdate(sourcePath string, destinationPath string) error {
	u.log.Info("Checking update for %s", destinationPath)
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return err
	}

	destFileInfo, err := os.Lstat(destinationPath)
	if os.IsNotExist(err) {
		u.log.Info("Existing destination doesn't exist")
		return nil
	}

	if err != nil {
		return err
	}
	// Make sure destination is not a link
	if destFileInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("Destination is a symbolic link")
	}
	if !destFileInfo.Mode().IsDir() {
		return fmt.Errorf("Destination is a directory")
	}

	return u.checkPlatformSpecificUpdate(sourcePath, destinationPath)
}

func (u *Updater) applyFile(src string, dest string) (err error) {
	if _, sterr := os.Stat(dest); sterr == nil {
		u.log.Info("Removing %s", dest)
		err = os.RemoveAll(dest)
		if err != nil {
			return
		}
	}

	u.log.Info("Moving (update) %s to %s", src, dest)
	err = os.Rename(src, dest)
	return
}

// Update checks, downloads and performs an update.
func (u *Updater) Update(ctx Context) (update *keybase1.Update, err error) {
	var skipped bool
	update, skipped, err = u.updateSingleFlight(ctx)
	// Retry if skipped via singleflight
	if skipped {
		update, _, err = u.updateSingleFlight(ctx)
	}
	return
}

func (u *Updater) updateSingleFlight(ctx Context) (*keybase1.Update, bool, error) {
	if u.cancelPrompt != nil {
		u.log.Info("Canceling update that was waiting on a prompt")
		u.cancelPrompt()
	}

	do := func() (interface{}, error) {
		return u.update(ctx)
	}
	any, cached, err := u.callGroup.Do("update", do)
	if cached {
		u.log.Info("There was an update already in progress")
		return nil, true, nil
	}

	update, ok := any.(*keybase1.Update)
	if !ok {
		return nil, false, fmt.Errorf("Invalid type returned by updater singleflight")
	}
	return update, false, err
}

func (u *Updater) update(ctx Context) (update *keybase1.Update, err error) {
	options, err := ctx.UpdateOptions()
	if err != nil {
		return
	}
	update, err = u.checkForUpdate(options, false)
	if err != nil {
		return
	}
	if update == nil {
		// No update available
		return
	}

	err = u.promptForUpdateAction(ctx, *update)
	if err != nil {
		return
	}

	// Linux updates don't have assets so it's ok to prompt for update before
	// asset checks.
	if update.Asset == nil {
		u.log.Info("No update asset to apply")
		return
	}
	if update.Asset.LocalPath == "" {
		err = fmt.Errorf("No local asset path for update")
		return
	}

	err = u.ensureAssetExists(*update.Asset)
	if err != nil {
		return
	}

	err = u.apply(ctx, options.DestinationPath, *update)
	if err != nil {
		return
	}

	_, err = u.restart(ctx)

	if update.Asset != nil {
		u.cleanup([]string{unzipDestination(update.Asset.LocalPath), update.Asset.LocalPath})
	}

	return
}

func (u *Updater) apply(ctx Context, destinationPath string, update keybase1.Update) (err error) {
	err = u.verifySignature(ctx, update)
	if err != nil {
		return
	}

	err = u.applyUpdate(update.Asset.LocalPath, destinationPath)
	return
}

func (u *Updater) updateUI(ctx Context) (updateUI UpdateUI, err error) {
	if ctx == nil {
		err = fmt.Errorf("No update UI available")
		return
	}

	return u.waitForUI(ctx, 5*time.Second)
}

func (u *Updater) promptForUpdateAction(ctx Context, update keybase1.Update) (err error) {

	u.log.Debug("+ Updater.promptForUpdateAction")
	defer func() {
		u.log.Debug("- Updater.promptForUpdateAction -> %v", err)
	}()

	auto, autoSet := u.config.GetUpdatePreferenceAuto()
	if auto {
		u.log.Debug("| going ahead with auto-updates")
		return
	}

	updateUI, err := u.updateUI(ctx)
	if err != nil {
		return
	}

	// If automatically apply not set, default to true
	alwaysAutoInstall := !autoSet

	promptOptions := keybase1.UpdatePromptOptions{
		AlwaysAutoInstall: alwaysAutoInstall,
	}

	_, canceler := context.WithCancel(context.Background())
	u.cancelPrompt = canceler
	updatePromptResponse, err := updateUI.UpdatePrompt(update, promptOptions)
	u.cancelPrompt = nil
	if err != nil {
		return
	}

	u.log.Debug("Update prompt response: %#v", updatePromptResponse)
	err = u.config.SetUpdatePreferenceAuto(updatePromptResponse.AlwaysAutoInstall)
	if err != nil {
		u.log.Warning("Error setting auto preference: %s", err)
	}
	switch updatePromptResponse.Action {
	case keybase1.UpdateActionPerformUpdate:
		// Continue
	case keybase1.UpdateActionSnooze:
		err = NewCanceledError("Snoozed update")
	case keybase1.UpdateActionCancel:
		err = NewCanceledError("Canceled by user")
	default:
		err = NewCanceledError("Canceled by service")
	}
	return
}

func (u *Updater) applyZip(localPath string, destinationPath string) (err error) {
	unzipPath, err := u.unpack(localPath)
	if err != nil {
		return
	}
	u.log.Info("Unzip path: %s", unzipPath)

	baseName := filepath.Base(destinationPath)
	sourcePath := filepath.Join(unzipPath, baseName)
	err = u.checkUpdate(sourcePath, destinationPath)
	if err != nil {
		return
	}

	err = u.applyFile(sourcePath, destinationPath)
	return
}

func (u *Updater) restart(ctx Context) (didQuit bool, err error) {
	return
}

func (u *Updater) cleanup(files []string) {
	u.log.Debug("Cleaning up after update")
	for _, f := range files {
		u.log.Debug("Removing %s", f)
		if f != "" {
			err := os.RemoveAll(f)
			if err != nil {
				u.log.Warning("Error trying to remove file (cleaning up after update): %s", err)
			}
		}
	}
}

func pathForUpdaterFilename(filename string) string {
	return filepath.Join(os.TempDir(), "KeybaseUpdates", filename)
}

func unzipDestination(filename string) string {
	return fmt.Sprintf("%s.unzipped", filename)
}

func copyFile(sourcePath string, destinationPath string) error {
	in, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	_, err = io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		return err
	}
	return cerr
}

func (u *Updater) verifySignature(ctx Context, update keybase1.Update) error {
	file, err := os.Open(update.Asset.LocalPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	if update.Asset.Signature == "" {
		u.log.Warning("No signature to verify")
		// TODO: Return error here to enable signature verification
		// return fmt.Errorf("No signature to verify")
		return nil
	}
	u.log.Info("Verifying signature %s", update.Asset.Signature)
	return ctx.Verify(file, update.Asset.Signature)
}

// waitForUI waits for a UI to be available. A UI might be missing for a few
// seconds between restarts.
func (u *Updater) waitForUI(ctx Context, wait time.Duration) (updateUI UpdateUI, err error) {
	t := time.Now()
	i := 1
	for time.Now().Sub(t) < wait {
		updateUI, err = ctx.GetUpdateUI()
		if err != nil || updateUI != nil {
			return
		}
		// Tell user we're waiting for UI after 4 seconds, every 4 seconds
		if i%4 == 0 {
			u.log.Info("Waiting for UI to be available...")
		}
		time.Sleep(time.Second)
		i++
	}
	return nil, fmt.Errorf("No UI available for updater")
}
