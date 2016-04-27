// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUpdateChecker(t *testing.T) {
	testServer := testServerForUpdateFile(t, testZipPath)
	defer testServer.Close()
	updater, err := newTestUpdaterWithServer(t, testServer)
	assert.NoError(t, err)

	checker := newUpdateChecker(updater, testUpdateCheckUI{promptDelay: 10 * time.Millisecond}, log, time.Millisecond)
	defer checker.Stop()
	checker.Start()

	time.Sleep(11 * time.Millisecond)

	assert.True(t, checker.Count() >= 1)
}

type testUpdateCheckUI struct {
	promptDelay time.Duration
}

func (u testUpdateCheckUI) UpdatePrompt(_ Update, _ UpdateOptions, _ UpdatePromptOptions) (*UpdatePromptResponse, error) {
	time.Sleep(u.promptDelay)
	return &UpdatePromptResponse{Action: UpdateActionApply}, nil
}

func (u testUpdateCheckUI) BeforeApply(update Update) error {
	return nil
}

func (u testUpdateCheckUI) AfterApply(update Update) error {
	return nil
}

func (u testUpdateCheckUI) GetUpdateUI() (UpdateUI, error) {
	return u, nil
}

func (u testUpdateCheckUI) Verify(update Update) error {
	return nil
}

func (u testUpdateCheckUI) Restart() error {
	return nil
}

func (u testUpdateCheckUI) UpdateOptions() UpdateOptions {
	return newDefaultTestUpdateOptions()
}

func (u testUpdateCheckUI) ReportError(_ Error, _ UpdateOptions) {}
