// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/keybase/go-updater"
	"github.com/stretchr/testify/assert"
)

type testConfigPlatform struct {
	config
}

func (c testConfigPlatform) promptPath() (string, error) {
	return filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/prompt-apply.sh"), nil
}

func TestUpdatePrompt(t *testing.T) {
	ctx := newContext(&testConfigPlatform{}, log)
	resp, err := ctx.UpdatePrompt(testUpdate, testOptions, updater.UpdatePromptOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
