// +build windows

package main

import (
	"encoding/json"
	"github.com/kardianos/osext"
	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater/command"
	"io/ioutil"
	"path/filepath"
	"time"
)

type updaterPromptInput struct {
	Title       string `json:"title"`
	Message     string `json:"message"`
	Description string `json:"description"`
	AutoUpdate  bool   `json:"autoUpdate"`
	OutPathName string `json:"outPathName"`
	TimeoutSecs int    `json:"timeoutSecs"`
}

func main() {
	var testLog = &logging.Logger{Module: "test"}
	//	testPrompt := []string{strings.Replace(`{"title":"Keybase Update: Version 1.0.16-20160603160930+43d4118","message":"The version you are currently running () is outdated.","description":"Please visit https://keybase.io for more information.","autoUpdate":true}`, "\\", "", -1)}
	exePathName, _ := osext.Executable()
	pathName, _ := filepath.Split(exePathName)
	outPathName := filepath.Join(pathName, "wintest_out.txt")

	promptJSONInput, err := json.Marshal(updaterPromptInput{
		Title:       "Keybase Update: Version Foobar",
		Message:     "The version you are currently running (0.0) is outdated.",
		Description: "Lots of cool stuff in here you need",
		AutoUpdate:  true,
		OutPathName: outPathName,
		TimeoutSecs: 10,
	})
	if err != nil {
		testLog.Errorf("Error generating input: %s", err)
		return
	}

	path := filepath.Join(pathName, "prompter", "prompter.hta")

	testLog.Debugf("Executing: %s", string(string(promptJSONInput)))

	_, err = command.Exec("mshta.exe", []string{path, string(promptJSONInput)}, 100*time.Second, testLog)
	if err != nil {
		testLog.Errorf("Error: %v", err)
		return
	}

	result, err := ioutil.ReadFile(outPathName)
	if err != nil {
		testLog.Errorf("Error opening result file: %v", err)
		return
	}

	testLog.Debugf("Result: %s", string(result))

}
