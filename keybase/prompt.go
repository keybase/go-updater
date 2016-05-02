// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/command"
)

type updaterPromptInput struct {
	Title       string `json:"title"`
	Message     string `json:"message"`
	Description string `json:"description"`
	AutoUpdate  bool   `json:"autoUpdate"`
}

type updaterPromptInputResult struct {
	Action     string `json:"action"`
	AutoUpdate bool   `json:"autoUpdate"`
}

// promptTimeout is a long timeout here cause it might show the prompt while the user is not present
var promptTimeout = time.Hour

func (c context) updatePrompt(promptCommand string, update updater.Update, options updater.UpdateOptions, promptOptions updater.UpdatePromptOptions) (*updater.UpdatePromptResponse, error) {
	description := update.Description
	if description == "" {
		description = "Please visit https://keybase.io for more information."
	}

	promptJSONInput, err := json.Marshal(updaterPromptInput{
		Title:       fmt.Sprintf("Keybase Update: Version %s", update.Version),
		Message:     fmt.Sprintf("The version you are currently running (%s) is outdated.", options.Version),
		Description: description,
		AutoUpdate:  promptOptions.AutoUpdate,
	})
	if err != nil {
		return nil, fmt.Errorf("Error generating input: %s", err)
	}

	var result updaterPromptInputResult
	if err := command.ExecForJSON(promptCommand, []string{string(promptJSONInput)}, &result, promptTimeout, c.log); err != nil {
		return nil, fmt.Errorf("Error running command: %s", err)
	}

	autoUpdate := false

	var updateAction updater.UpdateAction
	switch result.Action {
	case "apply":
		updateAction = updater.UpdateActionApply
		autoUpdate = result.AutoUpdate
	case "snooze":
		updateAction = updater.UpdateActionSnooze
	default:
		updateAction = updater.UpdateActionCancel
	}

	return &updater.UpdatePromptResponse{
		Action:     updateAction,
		AutoUpdate: autoUpdate,
	}, nil
}

type promptInput struct {
	Type    string   `json:"type"`
	Title   string   `json:"title"`
	Message string   `json:"message"`
	Buttons []string `json:"buttons"`
}

type promptInputResult struct {
	Button string `json:"button"`
}

func (c context) pausedPrompt(promptCommand string) error {
	promptJSONInput, err := json.Marshal(promptInput{
		Type:    "generic",
		Title:   "Update Paused",
		Message: "You have files, folders or a terminal open in Keybase.\n\nYou can force the update. That would be like yanking a USB drive and plugging it right back in. It'll instantly give you the latest version of Keybase, but you'll need to reopen any files you're working with. If you're working in the terminal, you'll need to cd out of /keybase and back in.",
		Buttons: []string{"Force update", "Try again later"},
	})
	if err != nil {
		return fmt.Errorf("Error generating input: %s", err)
	}

	var result promptInputResult
	if err := command.ExecForJSON(promptCommand, []string{string(promptJSONInput)}, &result, 5*time.Minute, c.log); err != nil {
		return fmt.Errorf("Error running command: %s", err)
	}

	switch result.Button {
	case "Force update":
		return nil
	default:
		return fmt.Errorf("Canceled by user (%s)", result.Button)
	}
}
