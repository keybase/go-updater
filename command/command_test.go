// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package command

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/keybase/go-logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logging.Logger{Module: "test"}

func TestExecEmpty(t *testing.T) {
	result, err := Exec("", nil, time.Second, log)
	assert.EqualError(t, err, "No command")
	assert.Equal(t, result.Stdout.String(), "")
	assert.Equal(t, result.Stderr.String(), "")
}

func TestExecInvalid(t *testing.T) {
	result, err := Exec("invalidexecutable", nil, time.Second, log)
	assert.EqualError(t, err, `exec: "invalidexecutable": executable file not found in $PATH`)
	assert.Equal(t, result.Stdout.String(), "")
	assert.Equal(t, result.Stderr.String(), "")
}

func TestExecEcho(t *testing.T) {
	result, err := Exec("echo", []string{"arg1", "arg2"}, time.Second, log)
	assert.NoError(t, err)
	assert.Equal(t, result.Stdout.String(), "arg1 arg2\n")
}

func TestExecNil(t *testing.T) {
	execCmd := func(name string, arg ...string) *exec.Cmd {
		return nil
	}
	_, err := execWithFunc("echo", []string{"arg1", "arg2"}, execCmd, time.Second, log)
	require.Error(t, err)
}

func TestExecTimeout(t *testing.T) {
	start := time.Now()
	timeout := 10 * time.Millisecond
	result, err := Exec("sleep", []string{"10"}, timeout, log)
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
	result, err := Exec("sleep", []string{"1"}, -time.Second, log)
	assert.Equal(t, result.Stdout.String(), "")
	assert.Equal(t, result.Stderr.String(), "")
	assert.EqualError(t, err, "Invalid timeout: -1s")
}

type testObj struct {
	StringVar string        `json:"stringVar"`
	NumberVar int           `json:"numberVar"`
	BoolVar   bool          `json:"boolVar"`
	ObjectVar testNestedObj `json:"objectVar"`
}

type testNestedObj struct {
	FloatVar float64 `json:"floatVar"`
}

const testJSON = `{
  "stringVar": "hi",
  "numberVar": 1,
  "boolVar": true,
  "objectVar": {
    "floatVar": 1.23
  }
}`

var testVal = testObj{
	StringVar: "hi",
	NumberVar: 1,
	BoolVar:   true,
	ObjectVar: testNestedObj{
		FloatVar: 1.23,
	},
}

func TestExecForJSON(t *testing.T) {
	var testValOut testObj
	err := ExecForJSON("echo", []string{testJSON}, &testValOut, time.Second, log)
	assert.NoError(t, err)
	t.Logf("Out: %#v", testValOut)
	if !reflect.DeepEqual(testVal, testValOut) {
		t.Errorf("Invalid object: %#v", testValOut)
	}
}

func TestExecForJSONEmpty(t *testing.T) {
	err := ExecForJSON("", nil, nil, time.Second, log)
	require.Error(t, err)
}

func TestExecForJSONInvalidObject(t *testing.T) {
	// Valid JSON, but not the right object
	validJSON := `{"stringVar": true}`
	var testValOut testObj
	err := ExecForJSON("echo", []string{validJSON}, &testValOut, time.Second, log)
	require.Error(t, err)
	t.Logf("Error: %s", err)
}

// TestExecForJSONAddingInvalidInput tests valid JSON input with invalid input after.
// We still succeed in this case since we got valid input to start.
func TestExecForJSONAddingInvalidInput(t *testing.T) {
	var testValOut testObj
	err := ExecForJSON("echo", []string{testJSON + "bad input"}, &testValOut, time.Second, log)
	assert.NoError(t, err)
	t.Logf("Out: %#v", testValOut)
	if !reflect.DeepEqual(testVal, testValOut) {
		t.Errorf("Invalid object: %#v", testValOut)
	}
}

func TestExecForJSONTimeout(t *testing.T) {
	var testValOut testObj
	err := ExecForJSON("sleep", []string{"10"}, &testValOut, 10*time.Millisecond, log)
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), "Error running command: timed out")
	}
}

// TestExecTimeoutProcessKilled checks to make sure process is killed after timeout
func TestExecTimeoutProcessKilled(t *testing.T) {
	result, err := execWithFunc("sleep", []string{"10"}, exec.Command, 10*time.Millisecond, log)
	assert.Equal(t, result.Stdout.String(), "")
	assert.Equal(t, result.Stderr.String(), "")
	assert.Error(t, err)
	require.NotNil(t, result.Process)
	findProcess, _ := os.FindProcess(result.Process.Pid)
	// This should error since killing a non-existant process should error
	perr := findProcess.Kill()
	assert.NotNil(t, perr, "Should have errored killing since killing non-existant process should error")
}

func TestExecNoExit(t *testing.T) {
	path := filepath.Join(os.Getenv("GOPATH"), "bin", "test")
	_, err := Exec(path, []string{}, 10*time.Millisecond, log)
	require.EqualError(t, err, "Error running command: timed out")
}

func TestExecOutput(t *testing.T) {
	testCommand := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/echo-out-err.sh")
	result, err := execWithFunc(testCommand, nil, exec.Command, time.Second, log)
	assert.NoError(t, err)
	assert.Equal(t, "stdout output\n", result.Stdout.String())
	assert.Equal(t, "stderr output\n", result.Stderr.String())
}