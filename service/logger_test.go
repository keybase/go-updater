// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"testing"

	"github.com/keybase/go-updater/util"
	"github.com/stretchr/testify/require"
)

func TestLoggerNil(t *testing.T) {
	log := logger{}
	log.Debug(nil)
	log.Debugf("", nil)
	log.Info(nil)
	log.Infof("", nil)
	log.Warning(nil)
	log.Warningf("", nil)
	log.Error(nil)
	log.Errorf("")
}

func TestLoggerFile(t *testing.T) {
	log := logger{}
	_, path, err := log.setLogToFile("KeybaseTest", "TestLoggerFile.log")
	defer util.RemoveFileAtPath(path)
	require.NoError(t, err)
	log.Debug("test")
}
