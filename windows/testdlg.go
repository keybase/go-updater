// +build windows

package main

import (
	"encoding/json"
	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater/command"
	"os"
	"path/filepath"
	"time"
)

type updaterPromptInput struct {
	Title       string `json:"title"`
	Message     string `json:"message"`
	Description string `json:"description"`
	AutoUpdate  bool   `json:"autoUpdate"`
}

func main() {
	var testLog = &logging.Logger{Module: "test"}
	//	testPrompt := []string{strings.Replace(`{"title":"Keybase Update: Version 1.0.16-20160603160930+43d4118","message":"The version you are currently running () is outdated.","description":"Please visit https://keybase.io for more information.","autoUpdate":true}`, "\\", "", -1)}

	promptJSONInput, err := json.Marshal(updaterPromptInput{
		Title:       "Keybase Update: Version Foobar",
		Message:     "The version you are currently running (0.0) is outdated.",
		Description: "Lots of cool stuff in here you need",
		AutoUpdate:  true,
	})
	if err != nil {
		testLog.Errorf("Error generating input: %s", err)
		return
	}

	path := filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "keybase", "go-updater", "windows", "prompter", "prompter.hta")
	result, err := command.Exec(path, []string{string(promptJSONInput)}, 100*time.Second, testLog)
	if err != nil {
		testLog.Errorf("Error: %v", err)
	} else {
		testLog.Debugf("Result: %v", result)
	}

}
