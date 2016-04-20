// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater"
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
		fmt.Printf("%s\n", "0.0.0")
		return
	}

	svc := serviceFromFlags(f)
	ret := run(svc)
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

	// TODO: Construct updater

	return newService(log)
}

func run(s *service) int {
	return s.Run()
}
