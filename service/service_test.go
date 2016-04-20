// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"testing"

	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater/keybase"
	"github.com/stretchr/testify/assert"
)

var log = logging.Logger{Module: "test"}

func TestService(t *testing.T) {
	ctx, upd := keybase.NewUpdaterContext("keybase", log)
	svc := newService(upd, ctx, log)
	assert.NotNil(t, svc)

	// svc.Start()
}
