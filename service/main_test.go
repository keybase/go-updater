// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlags(t *testing.T) {
	f := flags{
		pathToKeybase: "keybase",
	}
	svc := serviceFromFlags(f)
	require.NotNil(t, svc)
}
