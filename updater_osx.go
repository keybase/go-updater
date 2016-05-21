// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build darwin

package updater

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/keybase/go-updater/command"
	"github.com/keybase/go-updater/util"
)

func (u *Updater) check(sourcePath string, destinationPath string) error {
	// Check to make sure the update source path is a real directory
	ok, err := util.IsDirReal(sourcePath)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("Source path isn't a directory")
	}
	return nil
}

func (u *Updater) platformApplyUpdate(update Update, options UpdateOptions) error {
	localPath := update.Asset.LocalPath
	destinationPath := options.DestinationPath

	// The file name we unzip over should match the (base) file in the destination path
	filename := filepath.Base(destinationPath)
	if err := util.UnzipOver(localPath, filename, destinationPath, u.check, u.log); err != nil {
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
