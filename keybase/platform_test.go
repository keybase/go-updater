// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"os"
	"path/filepath"
)

type testConfigPlatform struct {
	config
}

func (c testConfigPlatform) promptPath() (string, error) {
	return filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/prompt-apply.sh"), nil
}
