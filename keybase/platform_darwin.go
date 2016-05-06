// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/kardianos/osext"
	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/command"
	"github.com/keybase/go-updater/process"
)

// destinationPath returns the app bundle path where this executable is located
func (c config) execPath() string {
	path, err := osext.Executable()
	if err != nil {
		c.log.Warningf("Error trying to determine our executable path: %s", err)
		return ""
	}
	return path
}

// destinationPath returns the app bundle path where this executable is located
func (c config) destinationPath() string {
	return appBundleForPath(c.execPath())
}

func appBundleForPath(path string) string {
	if path == "" {
		return ""
	}
	paths := strings.SplitN(path, ".app", 2)
	// If no match, return ""
	if len(paths) <= 1 {
		return ""
	}
	return paths[0] + ".app"
}

// dir returns directory for configuration files on darwin
func (c config) dir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	if c.appName == "" {
		return "", fmt.Errorf("Unable to resolve config directory: No app name")
	}
	return filepath.Join(usr.HomeDir, "Library", "Application Support", c.appName), nil
}

func (c config) osVersion() string {
	result, err := command.Exec("/usr/bin/sw_vers", []string{"-productVersion"}, 5*time.Second, c.log)
	if err != nil {
		c.log.Warningf("Error trying to determine OS version: %s (%s)", err, result.CombinedOutput())
		return ""
	}
	return strings.TrimSpace(result.Stdout.String())
}

func (c config) promptPath() (string, error) {
	destinationPath := c.destinationPath()
	if destinationPath == "" {
		return "", fmt.Errorf("No destination path")
	}
	return filepath.Join(destinationPath, "Contents", "Resources", "KeybaseUpdater.app", "Contents", "MacOS", "Updater"), nil
}

// UpdatePrompt is called when the user needs to accept an update
func (c context) UpdatePrompt(update updater.Update, options updater.UpdateOptions, promptOptions updater.UpdatePromptOptions) (*updater.UpdatePromptResponse, error) {
	promptPath, err := c.config.promptPath()
	if err != nil {
		return nil, err
	}
	return c.updatePrompt(promptPath, update, options, promptOptions, time.Hour)
}

// PausedPrompt is called when the we can't update cause the app is in use.
// We return true if the use wants to cancel the update.
func (c context) PausedPrompt() bool {
	promptPath, err := c.config.promptPath()
	if err != nil {
		c.log.Warningf("Error trying to get prompt path: %s", err)
		return false
	}
	cancelUpdate, err := c.pausedPrompt(promptPath, 5*time.Minute)
	if err != nil {
		c.log.Warningf("Error in paused prompt: %s", err)
		return false
	}
	return cancelUpdate
}

func (c context) Restart() error {
	appPath := c.config.destinationPath()
	if appPath == "" {
		return fmt.Errorf("No destination path for restart")
	}

	procName := filepath.Join(appPath, "Contents/MacOS/")
	process.TerminateAll(procName, time.Second, c.log)

	keybase := filepath.Join(appPath, "Contents/SharedSupport/bin/keybase")
	process.TerminateAll(keybase, time.Second, c.log)

	kbfs := filepath.Join(appPath, "Contents/SharedSupport/bin/kbfs")
	process.TerminateAll(kbfs, time.Second, c.log)

	if err := process.OpenAppDarwin(appPath, c.log); err != nil {
		c.log.Warningf("Error opening app: %s", err)
	}
	return nil
}
