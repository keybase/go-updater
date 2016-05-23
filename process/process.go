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

type matchFn func(ps.Process) bool

func matchPath(p ps.Process, match string, log logging.Logger) bool {
	path, err := p.Path()
	if err != nil {
		log.Warningf("Unable to get path for process: %s", err)
		return false
	}
	return path == match
}

func containsPath(p ps.Process, match string, log logging.Logger) bool {
	path, err := p.Path()
	if err != nil {
		log.Warningf("Unable to get path for process: %s", err)
		return false
	}
	return strings.Contains(path, match)
}

// FindProcesses returns processes containing string matching process path
func FindProcesses(match string, wait time.Duration, delay time.Duration, log logging.Logger) ([]ps.Process, error) {
	log.Infof("Finding process %q (max wait %s)", match, wait)
	matchPath := func(p ps.Process) bool { return containsPath(p, match, log) }
	start := time.Now()
	for i := 0; time.Since(start) < wait || i == 0; i++ {
		log.Debugf("Find process %q (%s < %s)", match, time.Since(start), wait)
		procs, err := findProcessesWithFn(ps.Processes, matchPath, 0)
		if err != nil {
			return nil, err
		}
		if len(procs) > 0 {
			return procs, nil
		}
		time.Sleep(delay)
	}
	return nil, nil
}

// findProcessWithPID with return process for a pid
func findProcessWithPID(pid int) (ps.Process, error) {
	matchPID := func(p ps.Process) bool { return p.Pid() == pid }
	procs, err := findProcessesWithFn(ps.Processes, matchPID, 1)
	if err != nil {
		return nil, err
	}
	if len(procs) == 0 {
		return nil, nil
	}
	return procs[0], nil
}

// Ignore deadcode warning
var _ = findProcessWithPID

// findProcessesWithFn finds processes using match function.
// If max is != 0, then we will return that max number of processes.
func findProcessesWithFn(processesFn processesFn, matchFn matchFn, max int) ([]ps.Process, error) {
	processes, err := processesFn()
	if err != nil {
		return nil, fmt.Errorf("Error listing processes: %s", err)
	}
	if processes == nil {
		return nil, nil
	}
	procs := []ps.Process{}
	for _, p := range processes {
		if matchFn(p) {
			procs = append(procs, p)
		}
		if max != 0 && len(procs) >= max {
			break
		}
	}
	return procs, nil
}

func findPIDsWithFn(fn processesFn, matchFn matchFn, log logging.Logger) ([]int, error) {
	procs, err := findProcessesWithFn(fn, matchFn, 0)
	if err != nil {
		return nil, err
	}
	pids := []int{}
	for _, p := range procs {
		pids = append(pids, p.Pid())
	}
	return pids, nil
}

// TerminateAll stops all processes with executable names that contains the matching string
func TerminateAll(match string, killDelay time.Duration, log logging.Logger) {
	terminateAll(ps.Processes, match, killDelay, log)
}

func terminateAll(fn processesFn, match string, killDelay time.Duration, log logging.Logger) {
	matchPath := func(p ps.Process) bool { return containsPath(p, match, log) }
	pids, err := findPIDsWithFn(fn, matchPath, log)
	if err != nil {
		log.Warningf("Error finding process: %s", err)
		return
	}
	if len(pids) == 0 {
		log.Warningf("No processes found matching %q", match)
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
