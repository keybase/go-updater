// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build !darwin,!windows

package updater

func (u *Updater) checkPlatformSpecificUpdate(sourcePath string, destinationPath string) error {
	return nil
}

func (u *Updater) applyUpdate(localPath string, destinationPath string) (err error) {
	return nil
}
