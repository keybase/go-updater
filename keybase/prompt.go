// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/keybase/go-updater"
)

type updaterPromptInput struct {
	Title       string `json:"title"`
	Message     string `json:"message"`
	Description string `json:"description"`
	AutoUpdate  bool   `json:"autoUpdate"`
}

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

	var result struct {
		Action     string `json:"action"`
		AutoUpdate bool   `json:"autoUpdate"`
	}

	// We have a long timeout here cause it might show the prompt while the user is not present
	if err := updater.RunJSONCommand(promptCommand, []string{string(promptJSONInput)}, &result, time.Hour, c.log); err != nil {
		return nil, fmt.Errorf("Error running command: %s", err)
	}

	var updateAction updater.UpdateAction
	switch result.Action {
	case "apply":
		updateAction = updater.UpdateActionApply
	case "snooze":
		updateAction = updater.UpdateActionSnooze
	default:
		updateAction = updater.UpdateActionCancel
	}

	return &updater.UpdatePromptResponse{
		Action:     updateAction,
		AutoUpdate: result.AutoUpdate,
	}, nil
}
