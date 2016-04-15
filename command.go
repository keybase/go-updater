// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/keybase/go-logging"
)

// RunCommand runs a command and returns the combined stdout/err output
func RunCommand(command string, args []string, timeout time.Duration, log logging.Logger) (string, error) {
	out, _, err := runCommand(command, args, true, timeout, log)
	if out == nil {
		return "", err
	}
	return string(out), err
}

// runCommand runs a command and returns the output. If combinedOutput is true,
// combined stdout/stderr output is returned, otherwise only stdout is returned.
// We will send TERM signal and wait 1 second or timeout, whichever is less,
// before calling KILL.
func runCommand(command string, args []string, combinedOutput bool, timeout time.Duration, log logging.Logger) ([]byte, *os.Process, error) {
	log.Debugf("Command: %s %s", command, args)
	if command == "" {
		return nil, nil, fmt.Errorf("No command")
	}
	if timeout < 0 {
		return nil, nil, fmt.Errorf("Invalid timeout: %s", timeout)
	}
	cmd := exec.Command(command, args...)
	if cmd == nil {
		return nil, nil, fmt.Errorf("No command")
	}
	// Run the command and spawn a goroutine to wait for it
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Start()
	if err != nil {
		return nil, nil, err
	}
	doneCh := make(chan error)
	go func() {
		doneCh <- cmd.Wait()
		close(doneCh)
	}()
	// Wait for the command to finish or time out
	select {
	case cmdErr := <-doneCh:
		return buf.Bytes(), cmd.Process, cmdErr
	case <-time.After(timeout):
		// Timed out
	}
	// If no process, nothing to kill
	if cmd.Process == nil {
		return buf.Bytes(), nil, fmt.Errorf("Error running command: no process")
	}

	// Signal the process to terminate gracefully
	// Wait a second or timeout for termination, whichever less
	termWait := time.Second
	if timeout < termWait {
		termWait = timeout
	}
	log.Warningf("Command timed out, terminating (will wait %s before killing)", termWait)
	cmd.Process.Signal(syscall.SIGTERM)
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
	return buf.Bytes(), cmd.Process, fmt.Errorf("Error running command: timed out")
}

// RunJSONCommand runs a command (with timeout) expecting JSON output with result interface
func RunJSONCommand(command string, args []string, result interface{}, timeout time.Duration, log logging.Logger) error {
	out, _, err := runCommand(command, args, false, timeout, log)
	if err != nil {
		return err
	}
	if out == nil {
		return fmt.Errorf("No output")
	}
	if err := json.NewDecoder(bytes.NewReader(out)).Decode(&result); err != nil {
		return fmt.Errorf("Error in result: %s", err)
	}
	return nil
}
