// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// Log is the logging interface for the command package
type Log interface {
	Debugf(s string, args ...interface{})
	Infof(s string, args ...interface{})
	Warningf(s string, args ...interface{})
	Errorf(s string, args ...interface{})
}

// Program is a program at path with arguments
type Program struct {
	Path string
	Args []string
}

// ArgsWith returns program args with passed in args
func (p Program) ArgsWith(args []string) []string {
	if p.Args == nil || len(p.Args) == 0 {
		return args
	}
	if len(args) == 0 {
		return p.Args
	}
	return append(p.Args, args...)
}

// Result is the result of running a command
type Result struct {
	Stdout  bytes.Buffer
	Stderr  bytes.Buffer
	Process *os.Process
}

// CombinedOutput returns Stdout and Stderr as a single string.
func (r Result) CombinedOutput() string {
	return fmt.Sprintf("[Stdout] %s\n[Stdrr] %s", r.Stdout.String(), r.Stderr.String())
}

type execCmd func(name string, arg ...string) *exec.Cmd

// Exec runs a command and returns the stdout/err output and error if any
func Exec(name string, args []string, timeout time.Duration, log Log) (Result, error) {
	return execWithFunc(name, args, exec.Command, timeout, log)
}

// ExecForJSON runs a command (with timeout) expecting JSON output with obj interface
func ExecForJSON(command string, args []string, obj interface{}, timeout time.Duration, log Log) error {
	result, err := execWithFunc(command, args, exec.Command, timeout, log)
	if err != nil {
		return err
	}
	if err := json.NewDecoder(&result.Stdout).Decode(&obj); err != nil {
		return fmt.Errorf("Error in result: %s", err)
	}
	return nil
}
