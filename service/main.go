// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/kardianos/osext"
	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/keybase"
	"github.com/keybase/go-updater/util"
)

type flags struct {
	version       bool
	logToFile     bool
	appName       string
	pathToKeybase string
	command       string
	ignoreSnooze  bool
}

func main() {
	f, args := loadFlags()
	if len(args) > 0 {
		f.command = args[0]
	}

	run(f)
}

func loadFlags() (flags, []string) {
	f := flags{}
	flag.BoolVar(&f.version, "version", false, "Show version")
	flag.BoolVar(&f.logToFile, "log-to-file", false, "Log to file")
	flag.StringVar(&f.pathToKeybase, "path-to-keybase", "", "Path to keybase executable")
	flag.StringVar(&f.appName, "app-name", defaultAppName(), "App name")
	flag.BoolVar(&f.ignoreSnooze, "ignore-snooze", true, "Ignore snooze, if not in service mode")
	flag.Parse()
	args := flag.Args()
	return f, args
}

func defaultAppName() string {
	if runtime.GOOS == "linux" {
		return "keybase"
	}
	return "Keybase"
}

func run(f flags) {
	if f.version {
		fmt.Printf("%s\n", updater.Version)
		return
	}
	ulog := logger{}

	if f.logToFile {
		logFile, _, err := ulog.setLogToFile(f.appName, "keybase.updater.log")
		if err != nil {
			ulog.Errorf("Error setting logging to file: %s", err)
		}
		defer util.Close(logFile)
	}

	// Set default path to keybase if not set
	if f.pathToKeybase == "" {
		path, err := osext.Executable()
		if err != nil {
			ulog.Warning("Error determining our executable path: %s", err)
		} else {
			dir, _ := filepath.Split(path)
			pathToKeybase := filepath.Join(dir, "keybase")
			ulog.Debugf("Using default path to keybase: %s", pathToKeybase)
			f.pathToKeybase = pathToKeybase
		}
	}

	if f.pathToKeybase == "" {
		ulog.Warning("Missing -path-to-keybase")
	}

	switch f.command {
	case "need-update":
		ctx, updater := keybase.NewUpdaterContext(f.appName, f.pathToKeybase, ulog, keybase.Check)
		needUpdate, err := updater.NeedUpdate(ctx)
		if err != nil {
			ulog.Error(err)
			os.Exit(1)
		}
		fmt.Println(needUpdate)
	case "check":
		if err := updateCheckFromFlags(f, ulog); err != nil {
			ulog.Error(err)
			os.Exit(1)
		}
	case "snooze":
		ctx, updater := keybase.NewUpdaterContext(f.appName, f.pathToKeybase, ulog, keybase.Check)
		if err := updater.Snooze(ctx); err != nil {
			ulog.Error(err)
			os.Exit(1)
		}
	case "service", "":
		svc := serviceFromFlags(f, ulog)
		svc.Run()
	case "clean":
		if runtime.GOOS == "windows" {
			ctx, _ := keybase.NewUpdaterContext(f.appName, f.pathToKeybase, ulog, keybase.CheckPassive)
			fmt.Printf("Doing DeepClean\n")
			ctx.DeepClean()
		} else {
			ulog.Errorf("Unknown command: %s", f.command)
		}
	default:
		ulog.Errorf("Unknown command: %s", f.command)
	}
}

func serviceFromFlags(f flags, ulog logger) *service {
	ulog.Infof("Updater %s", updater.Version)
	ctx, upd := keybase.NewUpdaterContext(f.appName, f.pathToKeybase, ulog, keybase.Service)
	return newService(upd, ctx, ulog, f.appName)
}

func updateCheckFromFlags(f flags, ulog logger) error {
	mode := keybase.CheckPassive
	if f.ignoreSnooze {
		mode = keybase.Check
	}
	ctx, updater := keybase.NewUpdaterContext(f.appName, f.pathToKeybase, ulog, mode)
	_, err := updater.Update(ctx)
	return err
}
