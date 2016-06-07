// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/command"
)

func (c config) destinationPath() string {
	// No destination path for Windows
	return ""
}

// Dir returns where to store config and log files
func Dir(appName string) (string, error) {
	dir := os.Getenv("APPDATA")
	if dir == "" {
		return "", fmt.Errorf("No APPDATA env set")
	}
	if appName == "" {
		return "", fmt.Errorf("No app name for dir")
	}
	return filepath.Join(dir, appName), nil
}

// LogDir is where to log
func LogDir(appName string) (string, error) {
	return Dir(appName)
}

func (c config) osVersion() string {
	result, err := command.Exec("cmd", []string{"/c", "ver"}, 5*time.Second, c.log)
	if err != nil {
		c.log.Warningf("Error trying to determine OS version: %s (%s)", err, result.CombinedOutput())
		return ""
	}
	return strings.TrimSpace(result.Stdout.String())
}

func (c config) promptProgram() (command.Program, error) {
	return command.Program{}, fmt.Errorf("Unsupported")
}

func (c config) notifyProgram() string {
	// No notify program for Windows
	return ""
}

func (c context) BeforeUpdatePrompt(update updater.Update, options updater.UpdateOptions) error {
	return nil
}

func (c context) UpdatePrompt(update updater.Update, options updater.UpdateOptions, promptOptions updater.UpdatePromptOptions) (*updater.UpdatePromptResponse, error) {
	// No update prompt for Windows, since the installer may handle it
	return &updater.UpdatePromptResponse{Action: updater.UpdateActionContinue}, nil
}

func (c context) PausedPrompt() bool {
	return false
}

func (c context) Apply(update updater.Update, options updater.UpdateOptions, tmpDir string) error {
	if update.Asset == nil || update.Asset.LocalPath == "" {
		return fmt.Errorf("No asset")
	}
	_, err := command.Exec(update.Asset.LocalPath, nil, time.Hour, c.log)
	return err
}

func (c context) Restart() error {
	// Restart is handled by the installer
	return nil
}
