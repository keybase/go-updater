// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/keybase/go-logging"
)

var log = logging.Logger{Module: "test"}

func newTestUpdater(t *testing.T) (*Updater, error) {
	return newTestUpdaterWithServer(t, nil)
}

func newTestUpdaterWithServer(t *testing.T, testServer *httptest.Server) (*Updater, error) {
	return NewUpdater(testUpdateSource{testServer: testServer}, &testConfig{}, log), nil
}

type testUpdateUI struct {
	options UpdateOptions
}

func (u testUpdateUI) UpdatePrompt(_ Update, _ UpdateOptions, _ UpdatePromptOptions) (*UpdatePromptResponse, error) {
	return &UpdatePromptResponse{Action: UpdateActionApply, AutoUpdate: true}, nil
}

func (u testUpdateUI) BeforeApply(update Update) error {
	return nil
}

func (u testUpdateUI) AfterApply(update Update) error {
	return nil
}

func (u testUpdateUI) GetUpdateUI() (UpdateUI, error) {
	return u, nil
}

func (u testUpdateUI) Verify(update Update) error {
	return SaltpackVerifyDetachedFileAtPath(update.Asset.LocalPath, update.Asset.Signature, validCodeSigningKIDs, log)
}

func (u testUpdateUI) Restart() error {
	return nil
}

func (u testUpdateUI) UpdateOptions() UpdateOptions {
	return u.options
}

type testUpdateSource struct {
	testServer *httptest.Server
}

func (u testUpdateSource) Description() string {
	return "Test"
}

func (u testUpdateSource) FindUpdate(config UpdateOptions) (*Update, error) {
	update := Update{
		Version:     "1.0.1",
		Name:        "Test",
		Description: "Bug fixes",
		InstallID:   "deadbeef",
		Asset: &Asset{
			Name:      "test.zip",
			URL:       u.testServer.URL,
			Digest:    "4769edbb33b86cf960cebd39af33cb3baabf38f8e43a8dc89ff420faf1cf0d36",
			Signature: testZipSignature,
		},
	}

	return &update, nil
}

type testConfig struct {
	auto      bool
	autoSet   bool
	installID string
}

func (c testConfig) GetUpdateAuto() (bool, bool) {
	return c.auto, c.autoSet
}

func (c *testConfig) SetUpdateAuto(b bool) error {
	c.auto = b
	c.autoSet = true
	return nil
}

func (c testConfig) GetInstallID() string {
	return c.installID
}

func (c *testConfig) SetInstallID(s string) error {
	c.installID = s
	return nil
}

func newDefaultTestUpdateOptions() UpdateOptions {
	return UpdateOptions{
		Version:         "1.0.0",
		Platform:        runtime.GOOS,
		DestinationPath: filepath.Join(os.TempDir(), "Test"),
	}
}
