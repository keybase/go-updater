// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package process

import (
	"testing"

	"github.com/keybase/go-logging"
	"github.com/stretchr/testify/assert"
)

var log = logging.Logger{Module: "test"}

func TestFindPIDs(t *testing.T) {
	pids, err := findPIDs("", log)
	assert.NoError(t, err)
	assert.True(t, len(pids) > 1)
}
