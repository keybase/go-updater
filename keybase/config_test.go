// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/util"
	"github.com/stretchr/testify/assert"
)

func testConfig(t *testing.T) (config, error) {
	testAppName, err := util.RandString("KeybaseTest", 20)
	if err != nil {
		t.Fatalf("Unable to resolve test app name: %s", err)
	}
	testPathToKeybase := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/keybase/keybase-test.sh")
	return newConfig(testAppName, testPathToKeybase, log)
}

func TestConfig(t *testing.T) {
	cfg, err := testConfig(t) // Will error since load fails on first newConfig
	assert.NotNil(t, err, "%s", err)
	path, err := cfg.path()
	assert.Nil(t, err, "%s", err)
	assert.NotEqual(t, path, "", "No config path")

	configDir, err := cfg.dir()
	assert.Nil(t, err, "%s", err)
	assert.NotEqual(t, configDir, "", "Config dir empty")
	defer util.RemoveFileAtPath(configDir)

	err = cfg.SetUpdateAuto(false)
	assert.Nil(t, err, "%s", err)
	auto, autoSet := cfg.GetUpdateAuto()
	assert.True(t, autoSet, "Auto should be set")
	assert.False(t, auto, "Auto should be false")
	err = cfg.SetUpdateAuto(true)
	auto, autoSet = cfg.GetUpdateAuto()
	assert.True(t, autoSet, "Auto should be set")
	assert.True(t, auto, "Auto should be true")

	err = cfg.SetInstallID("deadbeef")
	assert.Nil(t, err, "%s", err)
	assert.Equal(t, cfg.GetInstallID(), "deadbeef")

	err = cfg.save()
	assert.Nil(t, err, "%s", err)

	options := cfg.updaterOptions()
	t.Logf("Options: %#v", options)

	expectedOptions := updater.UpdateOptions{
		Version:         "1.2.3-400+cafebeef",
		Platform:        runtime.GOOS,
		DestinationPath: "",
		Channel:         "test",
		Env:             "prod",
		InstallID:       "deadbeef",
		Arch:            "amd64",
		Force:           false,
		OSVersion:       cfg.osVersion(),
		UpdaterVersion:  updater.Version,
	}

	assert.Equal(t, options, expectedOptions)

	// Load new config and make sure it has the same values
	cfg2, err := newConfig(cfg.appName, cfg.pathToKeybase, log)
	assert.Nil(t, err, "%s", err)
	assert.NotEqual(t, cfg2.path, "", "No config path")

	options2 := cfg2.updaterOptions()
	assert.Equal(t, options2, expectedOptions)

	auto2, autoSet2 := cfg2.GetUpdateAuto()
	assert.True(t, autoSet2, "Auto should be set")
	assert.True(t, auto2, "Auto should be true")
	assert.Equal(t, cfg2.GetInstallID(), "deadbeef")
}

func TestConfigBadPath(t *testing.T) {
	cfg := newDefaultConfig("", "", log)

	badPath := filepath.Join("/testdir", "updater.json") // Shouldn't be writable
	err := cfg.loadFromPath(badPath)
	t.Logf("Error: %#v", err)
	assert.NotNil(t, err, "Expected error")

	saveErr := cfg.saveToPath(badPath)
	t.Logf("Error: %#v", saveErr)
	assert.NotNil(t, saveErr, "Expected error")

	auto, autoSet := cfg.GetUpdateAuto()
	assert.False(t, autoSet, "Auto should not be set")
	assert.False(t, auto, "Auto should be false")
	assert.Equal(t, cfg.GetInstallID(), "")
}
