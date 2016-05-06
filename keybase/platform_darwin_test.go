// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build darwin

package keybase

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/keybase/go-updater/process"
	"github.com/stretchr/testify/assert"
)

func TestAppBundleForPath(t *testing.T) {
	assert.Equal(t, "", appBundleForPath(""))
	assert.Equal(t, "", appBundleForPath("foo"))
	assert.Equal(t, "/Applications/Keybase.app", appBundleForPath("/Applications/Keybase.app"))
	assert.Equal(t, "/Applications/Keybase.app", appBundleForPath("/Applications/Keybase.app/Contents/SharedSupport/bin/keybase"))
	assert.Equal(t, "/Applications/Keybase.app", appBundleForPath("/Applications/Keybase.app/Contents/Resources/Foo.app/Contents/MacOS/Foo"))
	assert.Equal(t, "", appBundleForPath("/Applications/Keybase.ap"))
	assert.Equal(t, "/Applications/Keybase.app", appBundleForPath("/Applications/Keybase.app/"))
}

type testConfigDarwin struct {
	testConfigPlatform
}

func (c testConfigDarwin) destinationPath() string {
	return filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/Test.app")
}

func TestRestart(t *testing.T) {
	ctx := newContext(&testConfigDarwin{}, log)
	err := ctx.Restart()
	defer process.TerminateAll(ctx.config.destinationPath(), 200*time.Millisecond, log)
	assert.NoError(t, err)
}
