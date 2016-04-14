// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/keybase/go-logging"
)

func timerWithTimeout(timeout time.Duration, cmd *exec.Cmd, log logging.Logger) *time.Timer {
	return time.AfterFunc(timeout, func() {
		if cmd != nil && cmd.Process != nil {
			log.Warningf("Command timed out, killing")
			if err := cmd.Process.Kill(); err != nil {
				log.Warningf("Error trying to kill process: %s", err)
			}
		}
	})
}

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
func runCommand(command string, args []string, combinedOutput bool, timeout time.Duration, log logging.Logger) ([]byte, *os.Process, error) {
	log.Debugf("Command: %s %s", command, args)
	if command == "" {
		return nil, nil, fmt.Errorf("No command")
	}
	cmd := exec.Command(command, args...)
	if cmd == nil {
		return nil, nil, fmt.Errorf("No command")
	}
	timer := timerWithTimeout(timeout, cmd, log)
	if timer != nil {
		defer timer.Stop()
	}
	var out []byte
	var err error
	if combinedOutput {
		// Both stdout and stderr
		out, err = cmd.CombinedOutput()
	} else {
		// Only stdout
		out, err = cmd.Output()
	}
	if err != nil {
		return out, cmd.Process, fmt.Errorf("Error running command: %s", err)
	}
	return out, cmd.Process, nil
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
