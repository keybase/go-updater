// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build !darwin,!windows

package updater

func (u *Updater) platformApplyUpdate(update Update, options UpdateOptions, tmpDir string) error {
	return nil
}
