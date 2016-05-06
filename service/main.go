// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/keybase"
)

type flags struct {
	version       bool
	pathToKeybase string
}

func main() {
	f := flags{}
	flag.BoolVar(&f.version, "version", false, "Show version")
	flag.StringVar(&f.pathToKeybase, "path-to-keybase", "", "Path to keybase executable")
	flag.Parse()

	if f.version {
		fmt.Printf("%s\n", updater.Version)
		return
	}

	svc := serviceFromFlags(f)
	ret := svc.Run()
	if ret != 0 {
		os.Exit(ret)
	}
}

func serviceFromFlags(f flags) *service {
	log := logging.Logger{Module: "service"}

	log.Infof("Updater %s", updater.Version)

	if f.pathToKeybase == "" {
		log.Warning("Missing -path-to-keybase")
	}

	ctx, upd := keybase.NewUpdaterContext(f.pathToKeybase, log)
	return newService(upd, ctx, log)
}
