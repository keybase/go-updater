// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build darwin

package process

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/keybase/go-logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logging.Logger{Module: "test"}

func TestRestartDarwin(t *testing.T) {
	appPath := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/Test.app")
	defer TerminateAll(appPath, log)

	err := OpenAppDarwin(appPath, log)
	require.NoError(t, err)

	err = RestartAppDarwin(appPath, log)
	require.NoError(t, err)
}

func TestFindPS(t *testing.T) {
	pids, err := findPS("ps", log)
	assert.NoError(t, err)
	require.Equal(t, 1, len(pids))
	assert.True(t, pids[0] != 0)
}

func TestParsePS(t *testing.T) {
	var ps = `
	  67846 /Applications/Keybase.app/Contents/SharedSupport/bin/keybase
		67847 /Applications/Keybase.app/Contents/SharedSupport/bin/keybase
      852 /Applications/Keybase.app/Contents/SharedSupport/bin/kbfs
        5 /Applications/Keybase.app/Contents/SharedSupport/bin/updater
       43 /Applications/Keybase.app/Contents/MacOS/Keybase
    67845 /Applications/Keybase.app/Contents/Frameworks/Keybase Helper.app/Contents/MacOS/Keybase Helper
      636  login ??
     1777 /usr/sbin/distnoted`
	pids, err := parsePS(bytes.NewBufferString(ps), "/Applications/Keybase.app/Contents/MacOS", log)
	assert.NoError(t, err)
	assert.Equal(t, []int{43}, pids)

	pids, err = parsePS(bytes.NewBufferString(ps), "/Applications/Keybase.app/Contents/SharedSupport/bin/keybase", log)
	assert.NoError(t, err)
	assert.Equal(t, []int{67846, 67847}, pids)

	pids, err = parsePS(bytes.NewBufferString(ps), "login", log)
	assert.NoError(t, err)
	assert.Equal(t, []int{636}, pids)
}

func TestParsePSNil(t *testing.T) {
	pids, err := parsePS(nil, "", log)
	require.EqualError(t, err, "Nothing to parse")
	assert.Nil(t, pids)
}

func TestParsePSNoPrefix(t *testing.T) {
	pids, err := parsePS(bytes.NewBuffer([]byte{}), "", log)
	require.EqualError(t, err, "No prefix")
	assert.Nil(t, pids)
}
