// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kardianos/osext"
	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/command"
	"math/rand"
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

type updaterWinPromptInput struct {
	Title       string `json:"title"`
	Message     string `json:"message"`
	Description string `json:"description"`
	AutoUpdate  bool   `json:"autoUpdate"`
	OutPathName string `json:"outPathName"`
	TimeoutSecs int    `json:"timeoutSecs"`
}

func (c context) UpdatePrompt(update updater.Update, options updater.UpdateOptions, promptOptions updater.UpdatePromptOptions) (*updater.UpdatePromptResponse, error) {
	promptProgram, err := c.config.promptProgram()
	if err != nil {
		return nil, err
	}

	promptJSONInput, err := c.promptInput(update, options, promptOptions)

	if err != nil {
		return nil, fmt.Errorf("Error generating input: %s", err)
	}

	// Unmarshal so we can add a couple of fields
	var promptArgs updaterPromptInput
	err = json.Unmarshal([]byte(promptJSONInput), &promptArgs)
	if err != nil {
		return nil, fmt.Errorf("Error generating input while unmarshaling: %s", err)
	}

	tmpPathname := filepath.Join(os.TempDir(), fmt.Sprintf("updatePrompt%d.txt", rand.Intn(1000)))
	defer os.Remove(tmpPathname)

	promptJSONWinInput, err := json.Marshal(updaterWinPromptInput{
		Title:       promptArgs.Title,
		Message:     promptArgs.Message,
		Description: promptArgs.Description,
		AutoUpdate:  promptArgs.AutoUpdate,
		OutPathName: tmpPathname,
		TimeoutSecs: 3600, // to match time.Hour, below
	})

	_, err = command.Exec(promptProgram.Path, promptProgram.ArgsWith([]string{string(promptJSONWinInput)}), time.Hour, c.log)
	if err != nil {
		return nil, fmt.Errorf("Error running command: %s", err)
	}

	result, err := c.updaterPromptResultFromFile(tmpPathname)
	if err != nil {
		return nil, err
	}
	return c.responseForResult(*result)
}

// updaterPromptResultFromFile gets the result from the temp file and decodes it
func (c context) updaterPromptResultFromFile(Name string) (*updaterPromptInputResult, error) {
	resultRaw, err := ioutil.ReadFile(Name)
	if err != nil {
		return nil, err
	}

	var result updaterPromptInputResult
	if err := json.Unmarshal(resultRaw, &result); err != nil {
		return nil, err
	}
	return &result, nil
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
