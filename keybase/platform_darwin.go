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
)

// destinationPath returns the app bundle path where this executable is located
func (c config) destinationPath() string {
	path, err := osext.Executable()
	if err != nil {
		c.log.Warningf("Error trying to determine our destination path: %s", err)
		return ""
	}
	return appBundleForPath(path)
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

func (c context) promptPath() (string, error) {
	destinationPath := c.config.destinationPath()
	if destinationPath == "" {
		return "", fmt.Errorf("No destination path")
	}
	return filepath.Join(destinationPath, "Contents", "Resources", "KeybaseUpdater.app", "Contents", "MacOS", "Updater"), nil
}

// UpdatePrompt is called when the user needs to accept an update
func (c context) UpdatePrompt(update updater.Update, options updater.UpdateOptions, promptOptions updater.UpdatePromptOptions) (*updater.UpdatePromptResponse, error) {
	promptPath, err := c.promptPath()
	if err != nil {
		return nil, err
	}
	return c.updatePrompt(promptPath, update, options, promptOptions)
}

// PausedPrompt is called when the we can't update cause the app is in use
func (c context) PausedPrompt() error {
	promptPath, err := c.promptPath()
	if err != nil {
		return err
	}
	return c.pausedPrompt(promptPath)
}
