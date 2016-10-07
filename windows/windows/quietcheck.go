// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build windows

package main

import (
	"flag"

	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater/keybase"
)

// Given the name of an installer, this can be run on a
// target system to see if it is going to upgrade Dokan.
func main() {
	dokanCode := flag.String("dokan", "", "DokanProductCode")
	var testLog = &logging.Logger{Module: "test"}

	isSilent, _ := keybase.CheckCanBeSilent(*dokanCode, *dokanCode, testLog, keybase.CheckRegistryUninstallCode)

	testLog.Debugf("Result: %v", isSilent)
}
