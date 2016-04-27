// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"fmt"
	"io"
	"net/http"

	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/keybase/go-logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logging.Logger{Module: "test"}

func newTestUpdater(t *testing.T) (*Updater, error) {
	return newTestUpdaterWithServer(t, nil)
}

func newTestUpdaterWithServer(t *testing.T, testServer *httptest.Server) (*Updater, error) {
	return NewUpdater(testUpdateSource{testServer: testServer}, &testConfig{}, log), nil
}

func newTestContext(options UpdateOptions, action UpdateAction) *testUpdateUI {
	return &testUpdateUI{options: options, action: action}
}

type testUpdateUI struct {
	action      UpdateAction
	options     UpdateOptions
	errReported *Error
}

func (u testUpdateUI) UpdatePrompt(_ Update, _ UpdateOptions, _ UpdatePromptOptions) (*UpdatePromptResponse, error) {
	return &UpdatePromptResponse{Action: u.action, AutoUpdate: true}, nil
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

func (u *testUpdateUI) ReportError(err Error, options UpdateOptions) {
	u.errReported = &err
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

func testServerForUpdateFile(t *testing.T, path string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open(path)
		require.NoError(t, err)
		w.Header().Set("Content-Type", "application/zip")
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	}))
}

func testServerForError(t *testing.T, err error) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, err.Error(), 500)
	}))
}

func TestUpdaterApply(t *testing.T) {
	testServer := testServerForUpdateFile(t, testZipPath)
	defer testServer.Close()

	upr, err := newTestUpdaterWithServer(t, testServer)
	assert.NoError(t, err)
	update, err := upr.Update(newTestContext(newDefaultTestUpdateOptions(), UpdateActionApply))
	require.NoError(t, err)
	require.NotNil(t, update)
	t.Logf("Update: %#v\n", *update)
	require.NotNil(t, update.Asset)
	t.Logf("Asset: %#v\n", *update.Asset)

	auto, autoSet := upr.config.GetUpdateAuto()
	assert.True(t, auto)
	assert.True(t, autoSet)
	assert.Equal(t, "deadbeef", upr.config.GetInstallID())
}

func TestUpdaterDownloadError(t *testing.T) {
	testServer := testServerForError(t, fmt.Errorf("bad response"))
	defer testServer.Close()

	upr, err := newTestUpdaterWithServer(t, testServer)
	assert.NoError(t, err)
	ctx := newTestContext(newDefaultTestUpdateOptions(), UpdateActionApply)
	_, err = upr.Update(ctx)
	assert.EqualError(t, err, "Update Error (download): Responded with 500 Internal Server Error")

	require.NotNil(t, ctx.errReported)
	assert.Equal(t, ctx.errReported.errorType, DownloadError)
}

func TestUpdaterCancel(t *testing.T) {
	testServer := testServerForError(t, fmt.Errorf("cancel"))
	defer testServer.Close()

	upr, err := newTestUpdaterWithServer(t, testServer)
	assert.NoError(t, err)
	ctx := newTestContext(newDefaultTestUpdateOptions(), UpdateActionCancel)
	_, err = upr.Update(ctx)
	assert.EqualError(t, err, "Update Error (cancel): Canceled by user")
}

func TestUpdaterSnooze(t *testing.T) {
	testServer := testServerForError(t, fmt.Errorf("snooze"))
	defer testServer.Close()

	upr, err := newTestUpdaterWithServer(t, testServer)
	assert.NoError(t, err)
	ctx := newTestContext(newDefaultTestUpdateOptions(), UpdateActionSnooze)
	_, err = upr.Update(ctx)
	assert.EqualError(t, err, "Update Error (cancel): Snoozed update")
}
