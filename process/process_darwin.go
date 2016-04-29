// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package process

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater/command"
)

// RestartAppDarwin restarts an app. We will still call open if the kill fails.
func RestartAppDarwin(appPath string, log logging.Logger) error {
	if appPath == "" {
		return fmt.Errorf("No app path to restart")
	}
	procName := filepath.Join(appPath, "Contents/MacOS/")
	TerminateAll(procName, log)
	return OpenAppDarwin(appPath, log)
}

// findPS finds PIDs for processes with prefix using ps.
// The command `ps ax -o pid,comm` returns process list in 2 columns, pid and executable name.
// For example:
//
//     67846 /Applications/Keybase.app/Contents/SharedSupport/bin/keybase
//     67847 /Applications/Keybase.app/Contents/SharedSupport/bin/keybase
//     53852 /Applications/Keybase.app/Contents/SharedSupport/bin/kbfs
//      3915 /Applications/Keybase.app/Contents/SharedSupport/bin/updater
//     67833 /Applications/Keybase.app/Contents/MacOS/Keybase
//
func findPS(prefix string, log logging.Logger) ([]int, error) {
	log.Debugf("Finding process with prefix: %q", prefix)
	result, err := command.Exec("ps", []string{"ax", "-o", "pid,comm"}, time.Minute, log)
	if err != nil {
		return nil, err
	}
	return parsePS(&result.Stdout, prefix, log)
}

func parsePS(reader io.Reader, prefix string, log logging.Logger) ([]int, error) {
	if reader == nil {
		return nil, fmt.Errorf("Nothing to parse")
	}
	if prefix == "" {
		return nil, fmt.Errorf("No prefix")
	}
	pids := []int{}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.HasPrefix(fields[1], prefix) {
			pid, err := strconv.Atoi(fields[0])
			if err != nil {
				log.Warningf("Invalid pid for %s", fields)
			} else if pid > 0 {
				pids = append(pids, pid)
			}
		}
	}
	return pids, nil
}

// TerminateAll stops processes with executable names that start with prefix
func TerminateAll(prefix string, log logging.Logger) {
	pids, err := findPS(prefix, log)
	if err != nil {
		log.Warningf("Error finding process: %s", err)
	}
	if pids == nil {
		log.Warningf("No processes found with prefix %q", prefix)
		return
	}
	for _, pid := range pids {
		if err := TerminatePid(pid, log); err != nil {
			log.Warningf("Error terminating %d: %s", pid, err)
		}
	}
}

// TerminatePid calls SIGTERM, then waits a second and then calls SIGKILL.
// We don't mind if we call SIGKILL on an already terminated process.
// If SIGKILL failed then we've got bigger problems.
func TerminatePid(pid int, log logging.Logger) error {
	log.Debugf("Searching OS for %d", pid)
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("Error finding OS process: %s", err)
	}
	if process == nil {
		return fmt.Errorf("No process found with pid %d", pid)
	}

	log.Debugf("Terminating: %#v", process)
	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		log.Warningf("Error sending terminate: %s", err)
	}
	time.Sleep(time.Second)
	_ = process.Kill()
	return nil
}

// OpenAppDarwin starts an app
func OpenAppDarwin(appPath string, log logging.Logger) error {
	tryOpen := func() error {
		result, err := command.Exec("/usr/bin/open", []string{appPath}, time.Minute, log)
		if err != nil {
			return fmt.Errorf("Open error: %s; %s", err, result.CombinedOutput())
		}
		return nil
	}
	// We need to try 10 times because Gatekeeper has some issues, for example,
	// http://www.openradar.me/23614087
	for i := 0; i < 10; i++ {
		err := tryOpen()
		if err == nil {
			break
		}
		log.Errorf("Open error (trying again in a second): %s", err)
		time.Sleep(1 * time.Second)
	}
	return nil
}
