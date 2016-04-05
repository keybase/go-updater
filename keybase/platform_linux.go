// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"os"
	"os/user"
	"path/filepath"
)

func destinationPath() (string, error) {
	return "", nil
}

func configDir(appName string) (string, error) {
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

func osVersion() string {
	return ""
}
