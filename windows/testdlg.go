// +build windows

package main

import (
	"encoding/json"
	"github.com/kardianos/osext"
	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater/command"
	"golang.org/x/sys/windows/registry"
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

	exePathName, _ := osext.Executable()
	pathName, _ := filepath.Split(exePathName)
	path := filepath.Join(pathName, "prompter", "prompter.hta")

	_, err = command.Exec("mshta.exe", []string{path, string(promptJSONInput)}, 100*time.Second, testLog)
	if err != nil {
		testLog.Errorf("Error: %v", err)
		return
	}

	registryKey, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Keybase`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		testLog.Errorf("Registry OpenKey error: %v", err)
		return
	}
	defer registryKey.Close()

	registryValue, _, err := registryKey.GetStringValue("UpdatePromptResult")
	if err != nil {
		testLog.Errorf("Error trying to read registry: %s", err)
		return
	}

	testLog.Debugf("Result: %s", registryValue)

}
