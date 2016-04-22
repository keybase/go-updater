// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build darwin

package updater

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"time"
)

func (u *Updater) checkPlatformSpecificUpdate(sourcePath string, destinationPath string) error {
	//
	// Get the uid, gid of the current user and make sure our src matches.
	//
	// Updating the modification time of the application is important because the
	// system will be aware a new version of your app is available.
	// Finder will report the correct file size and other metadata for it, URL
	// schemes your app may register will be updated, etc.
	//
	// This might fail if the app is owned by root/admin, in which case we should
	// get the priviledged helper tool involved.
	//

	// Get uid, gid of current user
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	uid, err := strconv.Atoi(currentUser.Uid)
	if err != nil {
		return err
	}
	gid, err := strconv.Atoi(currentUser.Gid)
	if err != nil {
		return err
	}

	u.log.Info("Current user uid: %d, gid: %d", uid, gid)

	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		err = os.Chown(path, uid, gid)
		if err != nil {
			return err
		}
		t := time.Now()
		return os.Chtimes(path, t, t)
	}

	u.log.Info("Touching, chowning files in %s", sourcePath)
	return filepath.Walk(sourcePath, walk)
}

func (u *Updater) openApplication(applicationPath string) error {
	tryOpen := func() error {
		out, err := exec.Command("/usr/bin/open", applicationPath).CombinedOutput()
		if err != nil {
			return fmt.Errorf("Open error: %s; %s", err, string(out))
		}
		return nil
	}
	for i := 0; i < 10; i++ {
		err := tryOpen()
		if err == nil {
			break
		}
		u.log.Errorf("Open error (trying again in a second): %s", err)
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (u *Updater) platformApplyUpdate(update Update, options UpdateOptions) error {
	// TODO
	return nil
}
