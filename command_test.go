// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/keybase/go-logging"
)

var log = logging.Logger{Module: "test"}

func TestEmptyRunCommand(t *testing.T) {
	out, err := RunCommand("", nil, time.Second, log)
	if out != "" {
		t.Errorf("Unexpected output: %s", out)
	}
	t.Logf("Error: %s", err)
	if err == nil {
		t.Fatal("Should have errored")
	}
}

func TestInvalidRunCommand(t *testing.T) {
	out, err := RunCommand("invalidexecutable", nil, time.Second, log)
	if out != "" {
		t.Errorf("Unexpected output: %s", out)
	}
	t.Logf("Error: %s", err)
	if err == nil {
		t.Fatal("Should have errored")
	}
}

func TestRunCommandEcho(t *testing.T) {
	out, err := RunCommand("echo", []string{"arg1", "arg2"}, time.Second, log)
	if err != nil {
		t.Fatal(err)
	}
	expected := "arg1 arg2\n"
	if out != expected {
		t.Errorf("Unexpected output: %q != %q", out, expected)
	}
}

func TestRunCommandTimeout(t *testing.T) {
	start := time.Now()
	out, err := RunCommand("sleep", []string{"10"}, time.Second, log)
	elapsed := time.Since(start)
	t.Logf("We elapsed %s", elapsed)
	if elapsed < time.Second {
		t.Error("We didn't actually sleep more than a second")
	}
	if out != "" {
		t.Errorf("Unexpected output: %s", out)
	}
	if err == nil {
		t.Fatal("Expected timeout error")
	}
	if err.Error() != "Error running command: signal: killed" {
		t.Errorf("Expected signal killed error, got %#v", err)
	}
}

func TestRunCommandBadTimeout(t *testing.T) {
	out, err := RunCommand("sleep", []string{"1"}, -time.Second, log)
	if out != "" {
		t.Errorf("Unexpected output")
	}
	t.Logf("Error: %s", err)
	if err == nil {
		t.Errorf("Bad timeout should error")
	}
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
	if err != nil {
		t.Fatal(err)
	}
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
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Out: %#v", testValOut)
	if !reflect.DeepEqual(testVal, testValOut) {
		t.Errorf("Invalid object: %#v", testValOut)
	}
}

func TestRunJSONCommandTimeout(t *testing.T) {
	log := logging.Logger{Module: "test"}
	var testValOut testObj
	err := RunJSONCommand("sleep", []string{"10"}, &testValOut, time.Second, log)
	if err == nil {
		t.Fatal("Expected timeout error")
	}
	if err.Error() != "Error running command: signal: killed" {
		t.Errorf("Expected signal killed error, got %#v", err)
	}
}

// TestTimeoutProcessKilled checks to make sure process is killed after timeout
func TestTimeoutProcessKilled(t *testing.T) {
	out, process, err := runCommand("sleep", []string{"10"}, true, time.Second, log)
	if out == nil {
		t.Errorf("Expected empty output")
	}
	if err == nil {
		t.Fatal("Expected timeout error")
	}
	findProcess, _ := os.FindProcess(process.Pid)
	// This should error since killing a non-existant process should error
	perr := findProcess.Kill()
	if perr == nil {
		t.Errorf("Process should be killed already: %s", perr)
	}
}
