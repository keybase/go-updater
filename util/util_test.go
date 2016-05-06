// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package util

import (
	"strings"
	"testing"

	"github.com/keybase/go-logging"
)

var testLog = logging.Logger{Module: "test"}

func TestJoinPredicate(t *testing.T) {
	f := func(s string) bool { return strings.HasPrefix(s, "f") }
	s := JoinPredicate([]string{"foo", "bar", "faa"}, "-", f)
	if s != "foo-faa" {
		t.Errorf("Unexpected output: %s", s)
	}
}

func TestRandString(t *testing.T) {
	s, err := RandString("prefix", 20)
	t.Logf("Rand string: %s", s)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(s, "prefix") {
		t.Errorf("Invalid prefix: %s", s)
	}
	if len(s) != 38 {
		t.Errorf("Invalid length: %s (%d)", s, len(s))
	}
}
