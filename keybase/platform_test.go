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
}

func (c testConfigPlatform) promptProgram() (command.Program, error) {
	return command.Program{
		Path: filepath.Join(os.Getenv("GOPATH"), "bin", "test"),
		Args: []string{"echo", `{
				"action": "apply",
  			"autoUpdate": true
			}`},
	}, nil
}
