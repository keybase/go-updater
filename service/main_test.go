// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import "testing"

func TestRun(t *testing.T) {
	f := flags{
		pathToKeybase: "keybase",
	}
	svc := serviceFromFlags(f)
	if svc == nil {
		t.Fatal("No service")
	}
}
