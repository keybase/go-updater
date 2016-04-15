// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package util

import "testing"

func TestDiscardAndCloseBodyNil(t *testing.T) {
	err := DiscardAndCloseBody(nil)
	if err == nil {
		t.Fatal("Should have errored")
	}
}
