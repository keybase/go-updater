// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build windows

package process

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTerminateAll(t *testing.T) {
	path := filepath.Join(os.Getenv("GOPATH"), "bin", "test")
	start := func() int {
		cmd := exec.Command(path, "sleep")
		err := cmd.Start()
		require.NoError(t, err)
		require.NotNil(t, cmd.Process)
		return cmd.Process.Pid
	}

	pids := []int{}
	pids = append(pids, start())
	pids = append(pids, start())
	TerminateAll("test", time.Millisecond, testLog)
	assertTerminated(t, pids[0], "exit status 1")
	assertTerminated(t, pids[1], "exit status 1")
}
