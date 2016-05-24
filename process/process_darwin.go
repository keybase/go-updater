// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package process

import (
	"fmt"
	"time"

	"github.com/keybase/go-updater/command"
)

// OpenAppDarwin starts an app
func OpenAppDarwin(appPath string, log Log) error {
	return openAppDarwin("/usr/bin/open", appPath, time.Second, log)
}

func openAppDarwin(bin string, appPath string, retryDelay time.Duration, log Log) error {
	tryOpen := func() error {
		result, err := command.Exec(bin, []string{appPath}, time.Minute, log)
		if err != nil {
			return fmt.Errorf("Open error: %s; %s", err, result.CombinedOutput())
		}
		return nil
	}
	// We need to try 10 times because Gatekeeper has some issues, for example,
	// http://www.openradar.me/23614087
	var err error
	for i := 0; i < 10; i++ {
		err = tryOpen()
		if err == nil {
			break
		}
		log.Errorf("Open error (trying again in %s): %s", retryDelay, err)
		time.Sleep(retryDelay)
	}
	return err
}
