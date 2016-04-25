// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build windows

package updater

import (
	"fmt"
	"os/exec"
)

func (u *Updater) openApplication(applicationPath string) error {
	return fmt.Errorf("Open application not supported on this platform")
}

func (u *Updater) platformApplyUpdate(update Update, options UpdateOptions) error {
	if update.Asset == nil || update.Asset.LocalPath == "" {
		return fmt.Errorf("No asset")
	}
	return exec.Command(update.Asset.LocalPath, "/SILENT").Start()
}
