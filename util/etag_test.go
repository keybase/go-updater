// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEtag(t *testing.T) {
	data := []byte("test data\n")
	path, err := WriteTempFile("TestEtag", data, 0644)
	assert.NoError(t, err)
	defer func() { _ = os.Remove(path) }()

	etag, err := ComputeEtag(path)
	assert.NoError(t, err)
	assert.Equal(t, "39a870a194a787550b6b5d1f49629236", etag)
}

func TestEtagInvalidPath(t *testing.T) {
	etag, err := ComputeEtag("/tmp/invalidpath")
	t.Logf("Error: %#v", err)
	assert.Error(t, err)
	assert.Equal(t, "", etag)
}
