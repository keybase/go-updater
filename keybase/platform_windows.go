// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kardianos/osext"
	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/command"
	"golang.org/x/sys/windows/registry"
)

func (c config) destinationPath() string {
	pathName, err := osext.Executable()
	if err != nil {
		c.log.Warningf("Error trying to determine our executable path: %s", err)
		return ""
	}
	dir, _ := filepath.Split(pathName)
	return dir
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

func (c config) notifyProgram() string {
	// No notify program for Windows
	return ""
}

func (c context) BeforeUpdatePrompt(update updater.Update, options updater.UpdateOptions) error {
	return nil
}

func (c config) promptProgram() (command.Program, error) {
	destinationPath := c.destinationPath()
	if destinationPath == "" {
		return command.Program{}, fmt.Errorf("No destination path")
	}

	return command.Program{
		Path: "mshta.exe",
		Args: []string{filepath.Join(destinationPath, "prompter", "prompter.hta")},
	}, nil
}

const registryUpdatePromptKeyName = "UpdatePromptResult"

func (c context) UpdatePrompt(update updater.Update, options updater.UpdateOptions, promptOptions updater.UpdatePromptOptions) (*updater.UpdatePromptResponse, error) {
	promptProgram, err := c.config.promptProgram()
	if err != nil {
		return nil, err
	}
	return c.updatePromptForProgram(promptProgram, update, options, promptOptions)
}

func (c context) updatePromptForProgram(promptProgram command.Program, update updater.Update, options updater.UpdateOptions, promptOptions updater.UpdatePromptOptions) (*updater.UpdatePromptResponse, error) {
	// Clear the result value we expect to find in the registry
	c.clearRegistryKey(registryUpdatePromptKeyName)

	promptJSONInput, err := c.promptInput(update, options, promptOptions)
	if err != nil {
		return nil, fmt.Errorf("Error generating input: %s", err)
	}

	_, err = command.Exec(promptProgram.Path, promptProgram.ArgsWith([]string{promptJSONInput}), time.Hour, c.log)
	if err != nil {
		return nil, fmt.Errorf("Error running command: %s", err)
	}

	result, err := c.updaterPromptResultFromRegistry()
	if err != nil {
		return nil, err
	}
	return c.responseForResult(*result)
}

// promptResultForRegistry gets the result from the registry and decodes it
func (c context) updaterPromptResultFromRegistry() (*updaterPromptInputResult, error) {
	registryKey, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Keybase`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return nil, fmt.Errorf("Error opening registry key: %s", err)
	}
	defer registryKey.Close()

	registryValue, _, err := registryKey.GetStringValue(registryUpdatePromptKeyName)
	if err != nil {
		return nil, fmt.Errorf("Error getting registry value: %s", err)
	}
	registryKey.DeleteValue(registryUpdatePromptKeyName)
	c.log.Debugf("Registry value: %s", registryValue)
	var result updaterPromptInputResult
	if err := json.Unmarshal([]byte(registryValue), &result); err != nil {
		return nil, fmt.Errorf("Error unmarshalling registry value: %s", err)
	}
	return &result, nil
}

func (c context) clearRegistryKey(s string) {
	registryKey, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Keybase`, registry.SET_VALUE)
	if err == nil {
		registryKey.DeleteValue(s)
	}
	registryKey.Close()
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
