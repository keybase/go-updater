// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/keybase/client/go/logger"
	"github.com/keybase/go-updater"
)

type flags struct {
	version       bool
	debug         bool
	dryRun        bool
	pathToKeybase string
}

func main() {
	f := flags{}
	flag.BoolVar(&f.debug, "debug", true, "Show debug")
	flag.BoolVar(&f.version, "version", false, "Show version")
	flag.BoolVar(&f.dryRun, "dry-run", true, "Dry run")
	flag.StringVar(&f.pathToKeybase, "path-to-keybase", "", "Path to keybase executable")
	flag.Parse()

	if f.version {
		fmt.Printf("%s\n", updater.Version)
		return
	}

	ret := run(f)
	if ret != 0 {
		os.Exit(ret)
	}
}

func run(f flags) int {
	var log = logger.New("service")
	log.Configure("", f.debug, "")

	log.Info("Updater %s", updater.Version)

	if f.pathToKeybase == "" {
		log.Error("Must specify -path-to-keybase")
		return 1
	}

	destPath, err := destinationPath()
	if err != nil {
		log.Warning("Error trying to find destination path: %s", err)
	}
	cfgDir, err := configDir("Keybase")
	if err != nil {
		log.Warning("Error trying to find config path: %s", err)
	}

	cfg, err := newConfig(cfgDir, destPath, f.pathToKeybase, log)
	if err != nil {
		log.Warning("Error loading config: %s", err)
	}

	src := NewKeybaseUpdateSource(log)
	upd := updater.NewUpdater(src, cfg, log)
	ctx := newContext(upd, cfg, log)

	log.Info("Starting service")
	svc := newService(upd, ctx, log)
	return svc.Run()
}
