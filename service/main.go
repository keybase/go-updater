// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/keybase"
)

type flags struct {
	version       bool
	pathToKeybase string
	command       string
}

func main() {
	f := flags{}
	flag.BoolVar(&f.version, "version", false, "Show version")
	flag.StringVar(&f.pathToKeybase, "path-to-keybase", "", "Path to keybase executable")
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		f.command = args[0]
	}

	run(f)
}

func run(f flags) {
	if f.version {
		fmt.Printf("%s\n", updater.Version)
		return
	}

	switch f.command {
	case "check":
		if err := updateCheckFromFlags(f); err != nil {
			log.Fatal(err)
		}
	case "service", "":
		svc := serviceFromFlags(f)
		svc.Run()
	default:
		log.Fatalf("Unknown command: %s", f.command)
	}
}

func serviceFromFlags(f flags) *service {
	log := &logging.Logger{Module: "service"}

	log.Infof("Updater %s", updater.Version)

	if f.pathToKeybase == "" {
		log.Warning("Missing -path-to-keybase")
	}

	ctx, upd := keybase.NewUpdaterContext(f.pathToKeybase, log)
	return newService(upd, ctx, log)
}

func updateCheckFromFlags(f flags) error {
	log := &logging.Logger{Module: "client"}

	ctx, updater := keybase.NewUpdaterContext(f.pathToKeybase, log)
	_, err := updater.Update(ctx)
	return err
}
