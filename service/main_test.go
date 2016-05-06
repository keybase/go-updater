// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceFlags(t *testing.T) {
	f := flags{
		pathToKeybase: "keybase",
	}
	svc := serviceFromFlags(f)
	require.NotNil(t, svc)
}

func TestServiceFlagsEmpty(t *testing.T) {
	svc := serviceFromFlags(flags{})
	require.NotNil(t, svc)
}
