// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/keybase/go-updater"
)

func (c config) destinationPath() string {
	// No destination path for Windows
	return ""
}

func (c config) dir() (string, error) {
	dir := os.Getenv("APPDATA")
	if dir == "" {
		return "", fmt.Errorf("No APPDATA env set")
	}
	return dir, nil
}

func (c config) osVersion() string {
	out, err := updater.RunCommand("cmd", []string{"/c", "ver"}, 5*time.Second, c.log)
	if err != nil {
		c.log.Warningf("Error trying to determine OS version: %s (%s)", err, out)
		return ""
	}
	return strings.TrimSpace(out)
}
