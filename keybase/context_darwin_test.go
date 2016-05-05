// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build darwin

package keybase

import (
	"testing"

	"github.com/keybase/go-updater"
	"github.com/stretchr/testify/require"
)

func TestContextCheckInUse(t *testing.T) {
	// In use, force
	ctx := newContext(&testConfigPausedPrompt{inUse: true, force: true}, log)
	err := ctx.BeforeApply(updater.Update{})
	require.NoError(t, err)

	// Not in use
	ctx = newContext(&testConfigPausedPrompt{inUse: false}, log)
	err = ctx.BeforeApply(updater.Update{})
	require.NoError(t, err)

	// In use, user cancels
	ctx = newContext(&testConfigPausedPrompt{inUse: true, force: false}, log)
	err = ctx.BeforeApply(updater.Update{})
	require.EqualError(t, err, "Canceled by user from paused prompt")
}
