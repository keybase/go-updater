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
	Args        []string
}

func (c testConfigPlatform) promptProgram() (command.Program, error) {
	var programPath = c.ProgramPath
	var args = c.Args
	if programPath == "" {
		programPath = filepath.Join(os.Getenv("GOPATH"), "bin", "test")
	}
	if len(args) == 0 {
		args = []string{"echo", `{
				"action": "apply",
  			"autoUpdate": true
			}`}
	}

	return command.Program{
		Path: programPath,
		Args: args,
	}, nil
}

func (c testConfigPlatform) notifyProgram() string {
	return "echo"
}
