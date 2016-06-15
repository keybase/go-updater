// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build windows

package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/kardianos/osext"
	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater/command"
)

// Copied here since it is not exported from go-updater/keybase
type updaterPromptInput struct {
	Title       string `json:"title"`
	Message     string `json:"message"`
	Description string `json:"description"`
	AutoUpdate  bool   `json:"autoUpdate"`
	OutPath     string `json:"outPath"` // Used for windows instead of stdout
}

func main() {
	var testLog = &logging.Logger{Module: "test"}

	exePathName, _ := osext.Executable()
	pathName, _ := filepath.Split(exePathName)
	outPathName := filepath.Join(pathName, "out.txt")

	promptJSONInput, err := json.Marshal(updaterPromptInput{
		Title:       "Keybase Update: Version Foobar",
		Message:     "The version you are currently running (0.0) is outdated.",
		Description: "Lots of cool stuff in here you need",
		AutoUpdate:  true,
		OutPath:     outPathName,
	})
	if err != nil {
		testLog.Errorf("Error generating input: %s", err)
		return
	}

	path := filepath.Join(pathName, "prompter.hta")

	testLog.Debugf("Executing: %s", string(string(promptJSONInput)))

	_, err = command.Exec("mshta.exe", []string{path, string(promptJSONInput)}, 100*time.Second, testLog)
	if err != nil {
		testLog.Errorf("Error: %v", err)
		return
	}

	result, err := ioutil.ReadFile(outPathName)
	if err != nil {
		testLog.Errorf("Error opening result file: %v", err)
		return
	}

	testLog.Debugf("Result: %s", string(result))
}
