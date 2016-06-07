// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build !windows

package command

import (
	"os"
	"os/exec"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecEcho(t *testing.T) {
	result, err := Exec("echo", []string{"arg1", "arg2"}, time.Second, testLog)
	assert.NoError(t, err)
	assert.Equal(t, result.Stdout.String(), "arg1 arg2\n")
}

func TestExecNil(t *testing.T) {
	execCmd := func(name string, arg ...string) *exec.Cmd {
		return nil
	}
	_, err := execWithFunc("echo", []string{"arg1", "arg2"}, execCmd, time.Second, testLog)
	require.Error(t, err)
}

func TestExecTimeout(t *testing.T) {
	start := time.Now()
	timeout := 10 * time.Millisecond
	result, err := Exec("sleep", []string{"10"}, timeout, testLog)
	elapsed := time.Since(start)
	t.Logf("We elapsed %s", elapsed)
	if elapsed < timeout {
		t.Error("We didn't actually sleep more than a second")
	}
	assert.Equal(t, result.Stdout.String(), "")
	assert.Equal(t, result.Stderr.String(), "")
	require.EqualError(t, err, "Error running command: timed out")
}

func TestExecBadTimeout(t *testing.T) {
	result, err := Exec("sleep", []string{"1"}, -time.Second, testLog)
	assert.Equal(t, result.Stdout.String(), "")
	assert.Equal(t, result.Stderr.String(), "")
	assert.EqualError(t, err, "Invalid timeout: -1s")
}
func TestExecForJSON(t *testing.T) {
	var testValOut testObj
	err := ExecForJSON("echo", []string{testJSON}, &testValOut, time.Second, testLog)
	assert.NoError(t, err)
	t.Logf("Out: %#v", testValOut)
	if !reflect.DeepEqual(testVal, testValOut) {
		t.Errorf("Invalid object: %#v", testValOut)
	}
}
func TestExecForJSONInvalidObject(t *testing.T) {
	// Valid JSON, but not the right object
	validJSON := `{"stringVar": true}`
	var testValOut testObj
	err := ExecForJSON("echo", []string{validJSON}, &testValOut, time.Second, testLog)
	require.Error(t, err)
	t.Logf("Error: %s", err)
}

// TestExecForJSONAddingInvalidInput tests valid JSON input with invalid input after.
// We still succeed in this case since we got valid input to start.
func TestExecForJSONAddingInvalidInput(t *testing.T) {
	var testValOut testObj
	err := ExecForJSON("echo", []string{testJSON + "bad input"}, &testValOut, time.Second, testLog)
	assert.NoError(t, err)
	t.Logf("Out: %#v", testValOut)
	if !reflect.DeepEqual(testVal, testValOut) {
		t.Errorf("Invalid object: %#v", testValOut)
	}
}

func TestExecForJSONTimeout(t *testing.T) {
	var testValOut testObj
	err := ExecForJSON("sleep", []string{"10"}, &testValOut, 10*time.Millisecond, testLog)
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), "Error running command: timed out")
	}
}

// TestExecTimeoutProcessKilled checks to make sure process is killed after timeout
func TestExecTimeoutProcessKilled(t *testing.T) {
	result, err := execWithFunc("sleep", []string{"10"}, exec.Command, 10*time.Millisecond, testLog)
	assert.Equal(t, result.Stdout.String(), "")
	assert.Equal(t, result.Stderr.String(), "")
	assert.Error(t, err)
	require.NotNil(t, result.Process)
	findProcess, _ := os.FindProcess(result.Process.Pid)
	// This should error since killing a non-existant process should error
	perr := findProcess.Kill()
	assert.NotNil(t, perr, "Should have errored killing since killing non-existant process should error")
}
