// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"fmt"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/kardianos/osext"
)

func destinationPath() (string, error) {
	path, err := osext.Executable()
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(path, "/Applications/Keybase.app/") {
		return "/Applications/Keybase.app", nil
	}
	return "", fmt.Errorf("No destination path found")
}

func configDir(appName string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, "Library", "Application Support", appName), nil
}

func osVersion() string {
	out, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
