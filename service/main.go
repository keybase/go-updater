// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"flag"
	"os"

	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/keybase"
	"github.com/keybase/go-updater/util"
)

type flags struct {
	version       bool
	logToFile     bool
	pathToKeybase string
	command       string
}

func main() {
	f := flags{}
	flag.BoolVar(&f.version, "version", false, "Show version")
	flag.BoolVar(&f.logToFile, "log-to-file", false, "Log to file")
	flag.StringVar(&f.pathToKeybase, "path-to-keybase", "", "Path to keybase executable")
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		f.command = args[0]
	}

	run(f)
}

func run(f flags) {
	ulog := logger{}
	if f.logToFile {
		logFile, err := ulog.setLogToFile("Keybase", "keybase.updater.log")
		if err != nil {
			ulog.Errorf("Error setting logging to file: %s", err)
		}
		defer util.Close(logFile)
	}

	if f.version {
		ulog.Infof("%s\n", updater.Version)
		return
	}

	switch f.command {
	case "check":
		if err := updateCheckFromFlags(f, ulog); err != nil {
			ulog.Error(err)
			os.Exit(1)
		}
	case "service", "":
		svc := serviceFromFlags(f, ulog)
		svc.Run()
	default:
		ulog.Errorf("Unknown command: %s", f.command)
	}
}

func serviceFromFlags(f flags, ulog logger) *service {
	ulog.Infof("Updater %s", updater.Version)

	if f.pathToKeybase == "" {
		ulog.Warning("Missing -path-to-keybase")
	}

	ctx, upd := keybase.NewUpdaterContext(f.pathToKeybase, ulog)
	return newService(upd, ctx, ulog)
}

func updateCheckFromFlags(f flags, ulog logger) error {
	ctx, updater := keybase.NewUpdaterContext(f.pathToKeybase, ulog)
	_, err := updater.Update(ctx)
	return err
}
