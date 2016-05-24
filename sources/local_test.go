// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package sources

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = &logging.Logger{Module: "test"}

func TestLocalUpdateSource(t *testing.T) {
	path := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/test.zip")
	jsonPath := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/update.json")
	local := NewLocalUpdateSource(path, jsonPath, log)
	assert.Equal(t, local.Description(), "Local")

	update, err := local.FindUpdate(updater.UpdateOptions{})
	require.NoError(t, err)
	require.NotNil(t, update)
}
