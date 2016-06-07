// +build windows

package main

import (
	"golang.org/x/sys/windows/registry"
)

func writeToRegistry(s string) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Keybase`, registry.SET_VALUE)
	if err != nil {
		return
	}
	defer k.Close()

	k.SetStringValue("UpdatePromptResult", s)
}
