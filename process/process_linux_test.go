// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build linux

package process

import (
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTerminateAll(t *testing.T) {
	start := func() int {
		cmd := exec.Command("sleep", "10")
		err := cmd.Start()
		require.NoError(t, err)
		require.NotNil(t, cmd.Process)
		return cmd.Process.Pid
	}

	pids := []int{}
	pids = append(pids, start())
	pids = append(pids, start())
	TerminateAll("/bin/sleep", time.Millisecond, testLog)
	assertTerminated(t, pids[0], "signal: terminated")
	assertTerminated(t, pids[1], "signal: terminated")
}
