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

func TestEmptyRunCommand(t *testing.T) {
	log := logging.Logger{Module: "test"}
	_, err := RunCommand("", nil, time.Second, log)
	t.Logf("Error: %s", err)
	if err == nil {
		t.Fatal("Should have errored")
	}
}

func TestInvalidRunCommand(t *testing.T) {
	log := logging.Logger{Module: "test"}
	_, err := RunCommand("invalidexecutable", nil, time.Second, log)
	t.Logf("Error: %s", err)
	if err == nil {
		t.Fatal("Should have errored")
	}
}

func TestRunCommandEcho(t *testing.T) {
	log := logging.Logger{Module: "test"}
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
	log := logging.Logger{Module: "test"}
	_, err := RunCommand("sleep", []string{"10"}, time.Second, log)
	if err == nil {
		t.Fatal("Expected timeout error")
	}
	if err.Error() != "Error running command: signal: killed" {
		t.Errorf("Expected signal killed error, got %#v", err)
	}
}

type testObj struct {
	StringVar string `json:"stringVar"`
	NumberVar int    `json:"numberVar"`
	BoolVar   bool   `json:"boolVar"`
}

const testJSON = `{
  "stringVar": "hi",
  "numberVar": 1,
  "boolVar": true
}`

var testVal = testObj{
	StringVar: "hi",
	NumberVar: 1,
	BoolVar:   true,
}

func TestRunJSONCommand(t *testing.T) {
	log := logging.Logger{Module: "test"}
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
	log := logging.Logger{Module: "test"}
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
	if err.Error() != "Error in result: EOF" {
		t.Errorf("Expected EOF error, got %#v", err)
	}
}

// TestTimeoutProcessState checks to make sure process is killed after timeout
func TestTimeoutProcessState(t *testing.T) {
	log := logging.Logger{Module: "test"}
	_, process, err := runCommand("sleep", []string{"10"}, time.Second, log)
	if err == nil {
		t.Fatal("Expected timeout error")
	}
	findProcess, _ := os.FindProcess(process.Pid)
	perr := findProcess.Kill()
	if perr == nil {
		t.Errorf("Process should be killed already: %s", perr)
	}
}
