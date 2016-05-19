// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build darwin

package updater

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"time"

	"github.com/keybase/go-updater/command"
	"github.com/keybase/go-updater/util"
)

func (u *Updater) platformProcessUpdate(sourcePath string, destinationPath string) error {
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

	u.log.Infof("Current user uid: %d, gid: %d", uid, gid)

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

func (u *Updater) check(sourcePath string, destinationPath string) error {
	ok, err := util.IsDirReal(sourcePath)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("Source path isn't a directory")
	}
	return nil
}

func (u *Updater) platformApplyUpdate(update Update, options UpdateOptions, tmpDir string) error {
	localPath := update.Asset.LocalPath
	destinationPath := options.DestinationPath

	// The file name we unzip over should match the (base) file in the destination path
	filename := filepath.Base(destinationPath)
	if err := util.UnzipOver(localPath, filename, destinationPath, u.check, tmpDir, u.log); err != nil {
		return err
	}

	if err := u.platformProcessUpdate(localPath, destinationPath); err != nil {
		return err
	}

	// Update spotlight
	u.log.Debugf("Updating spotlight: %s", destinationPath)
	spotlightResult, spotLightErr := command.Exec("/usr/bin/mdimport", []string{destinationPath}, 5*time.Second, u.log)
	if spotLightErr != nil {
		u.log.Warningf("Error trying to update spotlight: %s, (%s)", spotLightErr, spotlightResult.CombinedOutput())
	}

	return nil
}
