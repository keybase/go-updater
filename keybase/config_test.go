// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/util"
	"github.com/stretchr/testify/assert"
)

func testConfig(t *testing.T) (config, error) {
	testPathToKeybase := filepath.Join(os.Getenv("GOPATH"), "bin", "test")
	return newConfig("KeybaseTest", testPathToKeybase, testLog)
}

func TestConfig(t *testing.T) {
	cfg, err := testConfig(t) // Will error since load fails on first newConfig
	assert.NotNil(t, err, "%s", err)
	path, err := cfg.path()
	assert.NoError(t, err)
	assert.NotEqual(t, path, "", "No config path")

	configDir, err := cfg.dir()
	defer util.RemoveFileAtPath(configDir)
	assert.NoError(t, err)
	assert.NotEqual(t, configDir, "", "Config dir empty")
	defer util.RemoveFileAtPath(configDir)

	err = cfg.SetUpdateAuto(false)
	assert.NoError(t, err)
	auto, autoSet := cfg.GetUpdateAuto()
	assert.True(t, autoSet, "Auto should be set")
	assert.False(t, auto, "Auto should be false")
	err = cfg.SetUpdateAuto(true)
	auto, autoSet = cfg.GetUpdateAuto()
	assert.True(t, autoSet, "Auto should be set")
	assert.True(t, auto, "Auto should be true")

	err = cfg.SetInstallID("deadbeef")
	assert.NoError(t, err)
	assert.Equal(t, cfg.GetInstallID(), "deadbeef")

	err = cfg.save()
	assert.NoError(t, err)

	options := cfg.updaterOptions()
	t.Logf("Options: %#v", options)

	expectedOptions := updater.UpdateOptions{
		Version:         "1.2.3-400+cafebeef",
		Platform:        runtime.GOOS,
		DestinationPath: "",
		Channel:         "test",
		Env:             "prod",
		Arch:            runtime.GOARCH,
		Force:           false,
		OSVersion:       cfg.osVersion(),
		UpdaterVersion:  updater.Version,
	}

	assert.Equal(t, options, expectedOptions)

	// Load new config and make sure it has the same values
	cfg2, err := newConfig(cfg.appName, cfg.pathToKeybase, testLog)
	assert.NoError(t, err)
	assert.NotEqual(t, cfg2.path, "", "No config path")

	options2 := cfg2.updaterOptions()
	assert.Equal(t, options2, expectedOptions)

	auto2, autoSet2 := cfg2.GetUpdateAuto()
	assert.True(t, autoSet2, "Auto should be set")
	assert.True(t, auto2, "Auto should be true")
	assert.Equal(t, cfg2.GetInstallID(), "deadbeef")
}

func TestConfigBadPath(t *testing.T) {
	cfg := newDefaultConfig("", "", testLog)

	var badPath string
	if runtime.GOOS == "windows" {
		badPath = `x:\updater.json` // Shouldn't be writable
	} else {
		badPath = filepath.Join("/testdir", "updater.json") // Shouldn't be writable
	}

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

func TestConfigExtra(t *testing.T) {
	data := `{
	"extra": "extrafield",
	"installId": "deadbeef",
	"auto": false,
	"autoSet": true
	}`
	path := filepath.Join(os.TempDir(), "TestConfigExtra")
	defer util.RemoveFileAtPath(path)
	err := ioutil.WriteFile(path, []byte(data), 0644)
	assert.NoError(t, err)

	cfg := newDefaultConfig("", "", testLog)
	err = cfg.loadFromPath(path)
	assert.NoError(t, err)

	t.Logf("Config: %#v", cfg.store)
	assert.Equal(t, cfg.GetInstallID(), "deadbeef")
	auto, autoSet := cfg.GetUpdateAuto()
	assert.False(t, auto)
	assert.True(t, autoSet)
}

// TestConfigPartial tests that if any parsing error occurs, we have the
// default config
func TestConfigPartial(t *testing.T) {
	data := `{
	"auto": true,
	"installId": 1
	}`
	path := filepath.Join(os.TempDir(), "TestConfigBadType")
	defer util.RemoveFileAtPath(path)
	err := ioutil.WriteFile(path, []byte(data), 0644)
	assert.NoError(t, err)

	cfg := newDefaultConfig("", "", testLog)
	err = cfg.loadFromPath(path)
	assert.Error(t, err)
	auto, autoSet := cfg.GetUpdateAuto()
	assert.False(t, auto)
	assert.False(t, autoSet)
}

func TestKeybaseVersionInvalid(t *testing.T) {
	testPathToKeybase := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/err.sh")
	cfg, _ := newConfig("KeybaseTest", testPathToKeybase, testLog)
	version := cfg.keybaseVersion()
	assert.Equal(t, "", version)
}
