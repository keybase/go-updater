// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package command

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/keybase/go-logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testLog = &logging.Logger{Module: "test"}

func TestExecEmpty(t *testing.T) {
	result, err := Exec("", nil, time.Second, testLog)
	assert.EqualError(t, err, "No command")
	assert.Equal(t, result.Stdout.String(), "")
	assert.Equal(t, result.Stderr.String(), "")
}

func TestExecInvalid(t *testing.T) {
	result, err := Exec("invalidexecutable", nil, time.Second, testLog)
	assert.Error(t, err)
	require.True(t, strings.HasPrefix(err.Error(), `exec: "invalidexecutable": executable file not found in `))
	assert.Equal(t, result.Stdout.String(), "")
	assert.Equal(t, result.Stderr.String(), "")
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

func TestExecForJSONEmpty(t *testing.T) {
	err := ExecForJSON("", nil, nil, time.Second, testLog)
	require.Error(t, err)
}

// TestExecNoExit runs a go binary called test from package go-updater/test,
// that should be installed prior to running the tests.
func TestExecNoExit(t *testing.T) {
	path := filepath.Join(os.Getenv("GOPATH"), "bin", "test")
	_, err := Exec(path, []string{"noexit"}, 10*time.Millisecond, testLog)
	require.EqualError(t, err, "Error running command: timed out")
}

func TestExecOutput(t *testing.T) {
	path := filepath.Join(os.Getenv("GOPATH"), "bin", "test")
	result, err := execWithFunc(path, []string{"output"}, exec.Command, time.Second, testLog)
	assert.NoError(t, err)
	assert.Equal(t, "stdout output\n", result.Stdout.String())
	assert.Equal(t, "stderr output\n", result.Stderr.String())
}

func TestProgramArgsWith(t *testing.T) {
	assert.Equal(t, []string(nil), Program{Args: nil}.ArgsWith(nil))
	assert.Equal(t, []string(nil), Program{Args: []string{}}.ArgsWith(nil))
	assert.Equal(t, []string{}, Program{Args: nil}.ArgsWith([]string{}))
	assert.Equal(t, []string{}, Program{Args: []string{}}.ArgsWith([]string{}))
	assert.Equal(t, []string{"1"}, Program{Args: []string{"1"}}.ArgsWith(nil))
	assert.Equal(t, []string{"1", "2"}, Program{Args: []string{"1"}}.ArgsWith([]string{"2"}))
	assert.Equal(t, []string{"2"}, Program{Args: []string{}}.ArgsWith([]string{"2"}))
}
