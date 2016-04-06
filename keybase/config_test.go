// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/keybase/client/go/logger"
)

func TestNew(t *testing.T) {
	cfgDir, err := ioutil.TempDir("", "TestNewConfig")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(cfgDir)
	destDir, err := ioutil.TempDir("", "TestNewApp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(destDir)

	log := logger.New("test")
	cfg, err := newConfig(cfgDir, destDir, "keybase", log)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.path == "" {
		t.Fatal("No config path")
	}
	defer os.Remove(cfg.path)
	err = cfg.SetUpdatePreferenceAuto(false)
	if err != nil {
		t.Fatal(err)
	}
}
