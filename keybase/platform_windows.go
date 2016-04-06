// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"fmt"
	"os"
)

func destinationPath() (string, error) {
	return "", nil
}

func configDir(appName string) (string, error) {
	dir := os.Getenv("APPDATA")
	if dir == "" {
		return "", fmt.Errorf("No APPDATA env set")
	}
	return dir, nil
}

func osVersion() string {
	return ""
}
