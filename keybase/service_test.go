// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/keybase/client/go/logger"
	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/sources"
)

func TestService(t *testing.T) {
	cfgDir, err := ioutil.TempDir("", "TestServiceConfig")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(cfgDir)
	destDir, err := ioutil.TempDir("", "TestServiceApp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(destDir)

	log := logger.New("test")
	log.Configure("", true, "")
	cfg, err := newConfig(cfgDir, destDir, "keybase", log)
	if err != nil {
		t.Fatal(err)
	}
	src := sources.NewRemoteUpdateSource("https://prerelease-test.keybase.io", log)
	upd := updater.NewUpdater(src, cfg, log)

	ctx := newContext(upd, cfg, log)
	svc := newService(upd, ctx, log)
	svc.Start()

	time.Sleep(time.Second)
	// TODO: Verify checks ran
}
