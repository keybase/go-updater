// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/keybase/go-updater/command"
	"github.com/keybase/go-updater/util"
	"github.com/stretchr/testify/require"
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

func TestApplyNoAsset(t *testing.T) {
	ctx := newContext(&testConfigPlatform{}, testLog)
	tmpDir, err := util.WriteTempDir("TestApplyNoAsset.", 0700)
	require.NoError(t, err)
	err = ctx.Apply(testUpdate, testOptions, tmpDir)
	require.NoError(t, err)
}
