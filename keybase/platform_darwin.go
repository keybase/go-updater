// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/kardianos/osext"
	"github.com/keybase/go-updater"
)

// destinationPath returns the app bundle path where this executable is located.
// Currently we only support `/Applications/Keybase.app`.
func (c config) destinationPath() string {
	path, err := osext.Executable()
	if err != nil {
		c.log.Warningf("No destination path: %s", err)
		return ""
	}
	if strings.HasPrefix(path, "/Applications/Keybase.app/") {
		return "/Applications/Keybase.app"
	}
	return ""
}

// dir returns directory for configuration files on darwin
func (c config) dir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	if c.appName == "" {
		return "", fmt.Errorf("Unable to resolve config directory: No app name")
	}
	return filepath.Join(usr.HomeDir, "Library", "Application Support", c.appName), nil
}

func (c config) osVersion() string {
	out, err := updater.RunCommand("/usr/bin/sw_vers", []string{"-productVersion"}, 5*time.Second, c.log)
	if err != nil {
		c.log.Warningf("Error trying to determine OS version: %s (%s)", err, out)
		return ""
	}
	return strings.TrimSpace(out)
}
