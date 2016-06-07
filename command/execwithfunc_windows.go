// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build windows

package command

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
	"strings"
	"syscall"
	"time"
)

// exec runs a command and returns a Result and error if any.
// We will send TERM signal and wait 1 second or timeout, whichever is less,
// before calling KILL.
func execWithFunc(name string, args []string, execCmd execCmd, timeout time.Duration, log Log) (Result, error) {
	var result Result
	isHta := false
	log.Debugf("Execute: %s %v", name, args)
	if name == "" {
		return result, fmt.Errorf("No command")
	}
	if timeout < 0 {
		return result, fmt.Errorf("Invalid timeout: %s", timeout)
	}

	if strings.HasSuffix(name, ".hta") {
		isHta = true
		// Clean the result value we expect to find in the registry
		k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Keybase`, registry.SET_VALUE)
		if err == nil {
			k.DeleteValue("UpdateResult")
		}
		k.Close()

		args = append([]string{name}, args...)
		name = "mshta.exe"
		log.Debugf("Detected HTA: launching mshta.exe %v", args)
	}
	cmd := execCmd(name, args...)
	if cmd == nil {
		return result, fmt.Errorf("No command")
	}
	cmd.Stdout = &result.Stdout
	cmd.Stderr = &result.Stderr
	err := cmd.Start()
	if err != nil {
		return result, err
	}
	result.Process = cmd.Process
	doneCh := make(chan error)
	go func() {
		doneCh <- cmd.Wait()
		close(doneCh)
	}()
	// Wait for the command to finish or time out
	select {
	case cmdErr := <-doneCh:
		if isHta {
			k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Keybase`, registry.QUERY_VALUE|registry.SET_VALUE)
			if err == nil {
				defer k.Close()

				s, _, err := k.GetStringValue("UpdateResult")
				if err == nil && s != "" {
					log.Debugf("Got registry value: %s", s)
					result.Stdout.WriteString(s)
					k.DeleteValue("UpdateResult")
				}
			} else {
				log.Debugf("Couldn't open registry key - error %v", err)
			}
		}
		return result, cmdErr
	case <-time.After(timeout):
		// Timed out
	}
	// If no process, nothing to kill
	if cmd.Process == nil {
		return result, fmt.Errorf("Error running command: no process")
	}

	// Signal the process to terminate gracefully
	// Wait a second or timeout for termination, whichever less
	termWait := time.Second
	if timeout < termWait {
		termWait = timeout
	}
	log.Warningf("Command timed out, terminating (will wait %s before killing)", termWait)
	err = cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		log.Warningf("Error sending terminate: %s", err)
	}
	select {
	case <-doneCh:
		log.Warningf("Terminated")
	case <-time.After(termWait):
		// Bring out the big guns
		log.Warningf("Command failed to terminate, killing")
		if err := cmd.Process.Kill(); err != nil {
			log.Warningf("Error trying to kill process: %s", err)
		} else {
			log.Warningf("Killed process")
		}
	}
	return result, fmt.Errorf("Error running command: timed out")
}
