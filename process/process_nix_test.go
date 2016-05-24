// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build linux darwin

package process

import (
	"os/exec"
	"testing"
	"time"

	"github.com/keybase/go-updater/util"
	"github.com/stretchr/testify/require"
)

func TestTerminateAll(t *testing.T) {
	procPath := procPath(t)
	defer util.RemoveFileAtPath(procPath)
	start := func() int {
		cmd := exec.Command(procPath, "10")
		err := cmd.Start()
		require.NoError(t, err)
		require.NotNil(t, cmd.Process)
		return cmd.Process.Pid
	}

	pids := []int{}
	pids = append(pids, start())
	pids = append(pids, start())
	matcher := NewMatcher(procPath, PathEqual, testLog)
	TerminateAll(matcher, time.Millisecond, testLog)
	assertTerminated(t, pids[0], "signal: terminated")
	assertTerminated(t, pids[1], "signal: terminated")
}
