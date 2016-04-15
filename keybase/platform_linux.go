// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel/go-updater"
)

func (c config) destinationPath() string {
	// No destination path for Linux
	return ""
}

func (c config) dir() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir != "" {
		return dir, nil
	}
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, ".config"), nil
}

func (c config) osVersion() string {
	out, err := updater.RunCommand("uname", []string{"-mrs"}, 5*time.Second, c.log)
	if err != nil {
		c.log.Warningf("Error trying to determine OS version: %s (%s)", err, out)
		return ""
	}
	return strings.TrimSpace(out)
}
