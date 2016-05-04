// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateChecker(t *testing.T) {
	testServer := testServerForUpdateFile(t, testZipPath)
	defer testServer.Close()
	updater, err := newTestUpdaterWithServer(t, testServer, testUpdate(testServer.URL))
	assert.NoError(t, err)

	checker := newUpdateChecker(updater, testUpdateCheckUI{promptDelay: 10 * time.Millisecond}, log, time.Millisecond)
	defer checker.Stop()
	started := checker.Start()
	require.True(t, started)
	started = checker.Start()
	require.False(t, started)

	time.Sleep(11 * time.Millisecond)

	assert.True(t, checker.Count() >= 1)
}

type testUpdateCheckUI struct {
	promptDelay time.Duration
	verifyError error
}

func (u testUpdateCheckUI) UpdatePrompt(_ Update, _ UpdateOptions, _ UpdatePromptOptions) (*UpdatePromptResponse, error) {
	if u.promptDelay > 0 {
		time.Sleep(u.promptDelay)
	}
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
	return u.verifyError
}

func (u testUpdateCheckUI) Restart() error {
	return nil
}

func (u testUpdateCheckUI) UpdateOptions() UpdateOptions {
	return newDefaultTestUpdateOptions()
}

func (u testUpdateCheckUI) ReportAction(_ UpdateAction, _ *Update, _ UpdateOptions) {}

func (u testUpdateCheckUI) ReportError(_ error, _ *Update, _ UpdateOptions) {}

func (u testUpdateCheckUI) ReportSuccess(_ *Update, _ UpdateOptions) {}

func TestUpdateCheckerError(t *testing.T) {
	testServer := testServerForUpdateFile(t, testZipPath)
	defer testServer.Close()
	updater, err := newTestUpdaterWithServer(t, testServer, testUpdate(testServer.URL))
	assert.NoError(t, err)

	checker := NewUpdateChecker(updater, testUpdateCheckUI{verifyError: fmt.Errorf("Test verify error")}, log)
	err = checker.check()
	require.Error(t, err)
}
