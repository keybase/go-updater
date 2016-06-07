// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/command"
)

func (c config) destinationPath() string {
	// No destination path for Linux
	return ""
}

// Dir returns where to store config and log files
func Dir(appName string) (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir != "" {
		return dir, nil
	}
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	if appName == "" {
		return "", fmt.Errorf("No app name for dir")
	}
	return filepath.Join(usr.HomeDir, ".config", appName), nil
}

// LogDir is where to log
func LogDir(appName string) (string, error) {
	dir := os.Getenv("XDG_CACHE_HOME")
	if dir != "" {
		return dir, nil
	}
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	if appName == "" {
		return "", fmt.Errorf("No app name for dir")
	}
	return filepath.Join(usr.HomeDir, ".cache", appName), nil
}

func (c config) osVersion() string {
	result, err := command.Exec("uname", []string{"-mrs"}, 5*time.Second, c.log)
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
	return "notify-send"
}

func (c context) BeforeUpdatePrompt(update updater.Update, options updater.UpdateOptions) error {
	notifyArgs := []string{
		"-i", "/usr/share/icons/hicolor/128x128/apps/keybase.png",
		fmt.Sprintf("New Keybase version: %s", update.Version),
		fmt.Sprintf("Please update Keybase using your system package manager."),
	}
	result, err := command.Exec("notify-send", notifyArgs, 5*time.Second, c.log)
	if err != nil {
		c.log.Warningf("Error running notify-send: %s (%s)", err, result.CombinedOutput())
	}
	c.ReportAction(updater.UpdateActionSnooze, &update, options)
	return updater.CancelErr(fmt.Errorf("Linux uses system package manager"))
}

func (c context) UpdatePrompt(update updater.Update, options updater.UpdateOptions, promptOptions updater.UpdatePromptOptions) (*updater.UpdatePromptResponse, error) {
	// No update prompt for Linux
	return &updater.UpdatePromptResponse{Action: updater.UpdateActionContinue}, nil
}

func (c context) PausedPrompt() bool {
	return false
}

func (c context) Apply(update updater.Update, options updater.UpdateOptions, tmpDir string) error {
	return nil
}

func (c context) Restart() error {
	// Restart is handled by the installer
	return fmt.Errorf("Unsupported")
}
