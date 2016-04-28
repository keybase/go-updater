// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package process

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater/command"
)

// RestartAppDarwin restarts an app. We will still call open if the kill fails.
func RestartAppDarwin(appPath string, log logging.Logger) error {
	if appPath == "" {
		return fmt.Errorf("No app path to restart")
	}
	procName := filepath.Join(appPath, "Contents/MacOS/")
	TerminateAll(procName, log)
	return OpenAppDarwin(appPath, log)
}

// OpenAppDarwin starts an app
func OpenAppDarwin(appPath string, log logging.Logger) error {
	tryOpen := func() error {
		result, err := command.Exec("/usr/bin/open", []string{appPath}, time.Minute, log)
		if err != nil {
			return fmt.Errorf("Open error: %s; %s", err, result.CombinedOutput())
		}
		return nil
	}
	// We need to try 10 times because Gatekeeper has some issues, for example,
	// http://www.openradar.me/23614087
	for i := 0; i < 10; i++ {
		err := tryOpen()
		if err == nil {
			break
		}
		log.Errorf("Open error (trying again in a second): %s", err)
		time.Sleep(1 * time.Second)
	}
	return nil
}
