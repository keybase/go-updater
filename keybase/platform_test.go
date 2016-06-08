// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"os"
	"path/filepath"

	"github.com/keybase/go-updater/command"
)

type testConfigPlatform struct {
	config
	ProgramPath string
	EchoCommand string
}

func (c testConfigPlatform) promptProgram() (command.Program, error) {
	var programPath = c.ProgramPath
	var echoCommand = c.EchoCommand
	if programPath == "" {
		programPath = filepath.Join(os.Getenv("GOPATH"), "bin", "test")
	}
	if echoCommand == "" {
		echoCommand = "echo"
	}

	return command.Program{
		Path: programPath,
		Args: []string{echoCommand, `{
				"action": "apply",
  			"autoUpdate": true
			}`},
	}, nil
}

func (c testConfigPlatform) notifyProgram() string {
	return "echo"
}
