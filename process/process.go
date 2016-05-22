// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package process

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/keybase/go-logging"
	"github.com/keybase/go-ps"
)

type processesFn func() ([]ps.Process, error)

func findPIDsWithFn(fn processesFn, prefix string, log logging.Logger) ([]int, error) {
	processes, err := fn()
	if err != nil {
		return nil, fmt.Errorf("Error listing processes: %s", err)
	}
	if processes == nil {
		return nil, nil
	}
	pids := []int{}
	for _, p := range processes {
		path, err := p.Path()
		if err != nil {
			log.Warningf("Unable to get path for process: %s (will use executable name)", err)
			path = p.Executable()
		}
		if strings.HasPrefix(path, prefix) {
			pids = append(pids, p.Pid())
		}
	}
	return pids, nil
}

// TerminateAll stops all processes with executable names that start with prefix
func TerminateAll(prefix string, killDelay time.Duration, log logging.Logger) {
	terminateAll(ps.Processes, prefix, killDelay, log)
}

func terminateAll(fn processesFn, prefix string, killDelay time.Duration, log logging.Logger) {
	pids, err := findPIDsWithFn(fn, prefix, log)
	if err != nil {
		log.Warningf("Error finding process: %s", err)
		return
	}
	if len(pids) == 0 {
		log.Warningf("No processes found with prefix %q", prefix)
		return
	}
	for _, pid := range pids {
		if err := TerminatePID(pid, killDelay, log); err != nil {
			log.Warningf("Error terminating %d: %s", pid, err)
		}
	}
}

// TerminatePID is an overly simple way to terminate a PID.
// On darwin and linux, it calls SIGTERM, then waits a killDelay and then calls
// SIGKILL. We don't mind if we call SIGKILL on an already terminated process,
// since there could be a race anyway where the process exits right after we
// check if it's still running but before the SIGKILL.
// The killDelay is not used on windows.
func TerminatePID(pid int, killDelay time.Duration, log logging.Logger) error {
	log.Debugf("Searching OS for %d", pid)
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("Error finding OS process: %s", err)
	}
	if process == nil {
		return fmt.Errorf("No process found with pid %d", pid)
	}

	// Sending SIGTERM is not supported on windows, so we can use process.Kill()
	if runtime.GOOS == "windows" {
		return process.Kill()
	}

	log.Debugf("Terminating: %#v", process)
	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		log.Warningf("Error sending terminate: %s", err)
	}
	time.Sleep(killDelay)
	// Ignore SIGKILL error since it will be that the process wasn't running if
	// the terminate above succeeded. If terminate didn't succeed above, then
	// this SIGKILL is a measure of last resort, and an error would signify that
	// something in the environment has gone terribly wrong.
	_ = process.Kill()
	return err
}
