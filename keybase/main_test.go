// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import "testing"

// TODO
func testRunUpdate(t *testing.T) {
	f := flags{}
	ret := run(f)
	if ret != 0 {
		t.Fatalf("Exited with code: %d", ret)
	}
}
