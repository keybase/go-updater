// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/keybase/go-logging"
	"github.com/stretchr/testify/assert"
)

var log = logging.Logger{Module: "test"}

func TestEmptyRunCommand(t *testing.T) {
	out, err := RunCommand("", nil, time.Second, log)
	assert.Equal(t, out, "", "Should have empty output")
	t.Logf("Error: %s", err)
	assert.NotNil(t, err, "Should have errored")
}

func TestInvalidRunCommand(t *testing.T) {
	out, err := RunCommand("invalidexecutable", nil, time.Second, log)
	assert.Equal(t, out, "", "Should have empty output")
	t.Logf("Error: %s", err)
	assert.NotNil(t, err, "%s", err)
}

func TestRunCommandEcho(t *testing.T) {
	out, err := RunCommand("echo", []string{"arg1", "arg2"}, time.Second, log)
	assert.Nil(t, err, "Should have errored")
	assert.Equal(t, out, "arg1 arg2\n")
}

func TestRunCommandTimeout(t *testing.T) {
	start := time.Now()
	timeout := 10 * time.Millisecond
	out, err := RunCommand("sleep", []string{"10"}, timeout, log)
	elapsed := time.Since(start)
	t.Logf("We elapsed %s", elapsed)
	if elapsed < timeout {
		t.Error("We didn't actually sleep more than a second")
	}
	assert.Equal(t, out, "", "Should have empty output")
	if assert.NotNil(t, err, "Should have errored") {
		assert.Equal(t, err.Error(), "Error running command: timed out")
	}
}

func TestRunCommandBadTimeout(t *testing.T) {
	out, err := RunCommand("sleep", []string{"1"}, -time.Second, log)
	assert.Equal(t, out, "", "Should have empty output")
	t.Logf("Error: %s", err)
	assert.NotNil(t, err, "Should have errored")
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

func TestRunJSONCommand(t *testing.T) {
	var testValOut testObj
	err := RunJSONCommand("echo", []string{testJSON}, &testValOut, time.Second, log)
	assert.Nil(t, err, "%s", err)
	t.Logf("Out: %#v", testValOut)
	if !reflect.DeepEqual(testVal, testValOut) {
		t.Errorf("Invalid object: %#v", testValOut)
	}
}

// TestRunJSONCommandAddingInvalidInput tests valid JSON input with invalid input after.
// We still succeed in this case since we got valid input to start.
func TestRunJSONCommandAddingInvalidInput(t *testing.T) {
	var testValOut testObj
	err := RunJSONCommand("echo", []string{testJSON + "bad input"}, &testValOut, time.Second, log)
	assert.Nil(t, err, "%s", err)
	t.Logf("Out: %#v", testValOut)
	if !reflect.DeepEqual(testVal, testValOut) {
		t.Errorf("Invalid object: %#v", testValOut)
	}
}

func TestRunJSONCommandTimeout(t *testing.T) {
	log := logging.Logger{Module: "test"}
	var testValOut testObj
	err := RunJSONCommand("sleep", []string{"10"}, &testValOut, 10*time.Millisecond, log)
	if assert.NotNil(t, err, "Should have errored") {
		assert.Equal(t, err.Error(), "Error running command: timed out")
	}
}

// TestTimeoutProcessKilled checks to make sure process is killed after timeout
func TestTimeoutProcessKilled(t *testing.T) {
	out, process, err := runCommand("sleep", []string{"10"}, true, 10*time.Millisecond, log)
	assert.Equal(t, out, []byte{}, "Should have empty output")
	assert.NotNil(t, err, "Should have errored")
	findProcess, _ := os.FindProcess(process.Pid)
	// This should error since killing a non-existant process should error
	perr := findProcess.Kill()
	assert.NotNil(t, perr, "Should have errored killing since killing non-existant process should error")
}
