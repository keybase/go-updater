// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"io"
	"testing"
	"time"

	"github.com/keybase/client/go/logger"
	keybase1 "github.com/keybase/go-updater/protocol"
)

// TestUpdateCheckerIsAsync checks to make sure if the updater is blocked in a
// prompt that checks continue. This is safe because the updater is
// singleflighted.
func TestUpdateCheckerIsAsync(t *testing.T) {
	updater, err := newTestUpdater(t, nil)
	if err != nil {
		t.Fatal(err)
	}

	checker := newUpdateChecker(updater, testUpdateCheckUI{promptDelay: 400 * time.Millisecond}, logger.NewTestLogger(t), 100*time.Millisecond)
	defer checker.Stop()
	checker.Start()

	time.Sleep(400 * time.Millisecond)

	if checker.Count() <= 2 {
		t.Fatal("Checker should have checked more than once")
	}
}

type testUpdateCheckUI struct {
	promptDelay time.Duration
}

func (u testUpdateCheckUI) UpdatePrompt(_ keybase1.Update, _ keybase1.UpdatePromptOptions) (keybase1.UpdatePromptResponse, error) {
	time.Sleep(u.promptDelay)
	return keybase1.UpdatePromptResponse{Action: keybase1.UpdateActionPerformUpdate}, nil
}

func (u testUpdateCheckUI) GetUpdateUI() (UpdateUI, error) {
	return u, nil
}

func (u testUpdateCheckUI) Verify(r io.Reader, signature string) error {
	return nil
}

func (u testUpdateCheckUI) UpdateOptions() (keybase1.UpdateOptions, error) {
	return newDefaultTestUpdateOptions(), nil
}
