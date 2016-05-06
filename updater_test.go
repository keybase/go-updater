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
	return newTestUpdaterWithServer(t, nil, nil, &testConfig{})
}

func newTestUpdaterWithServer(t *testing.T, testServer *httptest.Server, update *Update, config Config) (*Updater, error) {
	return NewUpdater(testUpdateSource{testServer: testServer, update: update}, config, log), nil
}

func newTestContext(options UpdateOptions, cfg Config, action UpdateAction) *testUpdateUI {
	return &testUpdateUI{options: options, cfg: cfg, action: action}
}

type testUpdateUI struct {
	options            UpdateOptions
	cfg                Config
	action             UpdateAction
	promptErr          error
	verifyErr          error
	beforeApplyErr     error
	afterApplyErr      error
	restartErr         error
	errReported        error
	actionReported     UpdateAction
	autoUpdateReported bool
	updateReported     *Update
	successReported    bool
}

func (u testUpdateUI) UpdatePrompt(_ Update, _ UpdateOptions, _ UpdatePromptOptions) (*UpdatePromptResponse, error) {
	if u.promptErr != nil {
		return nil, u.promptErr
	}
	return &UpdatePromptResponse{Action: u.action, AutoUpdate: true}, nil
}

func (u testUpdateUI) BeforeApply(update Update) error {
	return u.beforeApplyErr
}

func (u testUpdateUI) AfterApply(update Update) error {
	return u.afterApplyErr
}

func (u testUpdateUI) GetUpdateUI() UpdateUI {
	return u
}

func (u testUpdateUI) Verify(update Update) error {
	if u.verifyErr != nil {
		return u.verifyErr
	}
	return SaltpackVerifyDetachedFileAtPath(update.Asset.LocalPath, update.Asset.Signature, validCodeSigningKIDs, log)
}

func (u testUpdateUI) Restart() error {
	return u.restartErr
}

func (u *testUpdateUI) ReportError(err error, update *Update, options UpdateOptions) {
	u.errReported = err
}

func (u *testUpdateUI) ReportAction(action UpdateAction, update *Update, options UpdateOptions) {
	u.actionReported = action
	autoUpdate, _ := u.cfg.GetUpdateAuto()
	u.autoUpdateReported = autoUpdate
	u.updateReported = update
}

func (u *testUpdateUI) ReportSuccess(update *Update, options UpdateOptions) {
	u.successReported = true
	u.updateReported = update
}

func (u testUpdateUI) UpdateOptions() UpdateOptions {
	return u.options
}

type testUpdateSource struct {
	testServer *httptest.Server
	update     *Update
	findErr    error
}

func (u testUpdateSource) Description() string {
	return "Test"
}

func testUpdate(uri string) *Update {
	return newTestUpdate(uri, true)
}

func newTestUpdate(uri string, needUpdate bool) *Update {
	update := &Update{
		Version:     "1.0.1",
		Name:        "Test",
		Description: "Bug fixes",
		InstallID:   "deadbeef",
		RequestID:   "cafedead",
		NeedUpdate:  needUpdate,
	}
	if uri != "" {
		update.Asset = &Asset{
			Name:      "test.zip",
			URL:       uri,
			Digest:    "3a147f31b25a6027bda15367def6f4499e29a9b531855c0ac881a8f3a83a12b9",
			Signature: testZipSignature,
		}
	}
	return update
}

func (u testUpdateSource) FindUpdate(options UpdateOptions) (*Update, error) {
	return u.update, u.findErr
}

type testConfig struct {
	auto      bool
	autoSet   bool
	installID string
	err       error
}

func (c testConfig) GetUpdateAuto() (bool, bool) {
	return c.auto, c.autoSet
}

func (c *testConfig) SetUpdateAuto(b bool) error {
	c.auto = b
	c.autoSet = true
	return c.err
}

func (c testConfig) GetInstallID() string {
	return c.installID
}

func (c *testConfig) SetInstallID(s string) error {
	c.installID = s
	return c.err
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

func testServerNotFound(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", 404)
	}))
}

func TestUpdaterApply(t *testing.T) {
	testServer := testServerForUpdateFile(t, testZipPath)
	defer testServer.Close()

	upr, err := newTestUpdaterWithServer(t, testServer, testUpdate(testServer.URL), &testConfig{})
	assert.NoError(t, err)
	ctx := newTestContext(newDefaultTestUpdateOptions(), upr.config, UpdateActionApply)
	update, err := upr.Update(ctx)
	require.NoError(t, err)
	require.NotNil(t, update)
	t.Logf("Update: %#v\n", *update)
	require.NotNil(t, update.Asset)
	t.Logf("Asset: %#v\n", *update.Asset)

	auto, autoSet := upr.config.GetUpdateAuto()
	assert.True(t, auto)
	assert.True(t, autoSet)
	assert.Equal(t, "deadbeef", upr.config.GetInstallID())

	assert.Nil(t, ctx.errReported)
	assert.Equal(t, ctx.actionReported, UpdateActionApply)
	assert.True(t, ctx.autoUpdateReported)

	require.NotNil(t, ctx.updateReported)
	assert.Equal(t, "deadbeef", ctx.updateReported.InstallID)
	assert.Equal(t, "cafedead", ctx.updateReported.RequestID)
	assert.True(t, ctx.successReported)

	assert.Equal(t, "apply", UpdateActionApply.String())
}

func TestUpdaterDownloadError(t *testing.T) {
	testServer := testServerForError(t, fmt.Errorf("bad response"))
	defer testServer.Close()

	upr, err := newTestUpdaterWithServer(t, testServer, testUpdate(testServer.URL), &testConfig{})
	assert.NoError(t, err)
	ctx := newTestContext(newDefaultTestUpdateOptions(), upr.config, UpdateActionApply)
	_, err = upr.Update(ctx)
	assert.EqualError(t, err, "Update Error (download): Responded with 500 Internal Server Error")

	require.NotNil(t, ctx.errReported)
	assert.Equal(t, ctx.errReported.(Error).errorType, DownloadError)
	assert.Equal(t, "deadbeef", ctx.updateReported.InstallID)
	assert.Equal(t, "cafedead", ctx.updateReported.RequestID)
	assert.False(t, ctx.successReported)
}

func TestUpdaterCancel(t *testing.T) {
	testServer := testServerForError(t, fmt.Errorf("cancel"))
	defer testServer.Close()

	upr, err := newTestUpdaterWithServer(t, testServer, testUpdate(testServer.URL), &testConfig{})
	assert.NoError(t, err)
	ctx := newTestContext(newDefaultTestUpdateOptions(), upr.config, UpdateActionCancel)
	_, err = upr.Update(ctx)
	assert.EqualError(t, err, "Update Error (cancel): Canceled")

	// Don't report error on user cancel
	assert.NoError(t, ctx.errReported)
}

func TestUpdaterSnooze(t *testing.T) {
	testServer := testServerForError(t, fmt.Errorf("snooze"))
	defer testServer.Close()

	upr, err := newTestUpdaterWithServer(t, testServer, testUpdate(testServer.URL), &testConfig{})
	assert.NoError(t, err)
	ctx := newTestContext(newDefaultTestUpdateOptions(), upr.config, UpdateActionSnooze)
	_, err = upr.Update(ctx)
	assert.EqualError(t, err, "Update Error (cancel): Snoozed update")

	// Don't report error on user snooze
	assert.NoError(t, ctx.errReported)
}

func TestUpdateNoAsset(t *testing.T) {
	testServer := testServerNotFound(t)
	defer testServer.Close()

	upr, err := newTestUpdaterWithServer(t, testServer, testUpdate(""), &testConfig{})
	assert.NoError(t, err)
	ctx := newTestContext(newDefaultTestUpdateOptions(), upr.config, UpdateActionApply)
	update, err := upr.Update(ctx)
	assert.NoError(t, err)
	assert.Nil(t, update.Asset)
}

func testUpdaterError(t *testing.T, errorType ErrorType) {
	testServer := testServerForUpdateFile(t, testZipPath)
	defer testServer.Close()

	upr, _ := newTestUpdaterWithServer(t, testServer, testUpdate(testServer.URL), &testConfig{})
	ctx := newTestContext(newDefaultTestUpdateOptions(), upr.config, UpdateActionApply)
	testErr := fmt.Errorf("Test error")
	switch errorType {
	case PromptError:
		ctx.promptErr = testErr
	case VerifyError:
		ctx.verifyErr = testErr
	case RestartError:
		ctx.restartErr = testErr
	}

	_, err := upr.Update(ctx)
	assert.EqualError(t, err, fmt.Sprintf("Update Error (%s): Test error", errorType.String()))

	require.NotNil(t, ctx.errReported)
	assert.Equal(t, ctx.errReported.(Error).errorType, errorType)
}

func TestUpdaterErrors(t *testing.T) {
	testUpdaterError(t, PromptError)
	testUpdaterError(t, VerifyError)
	testUpdaterError(t, RestartError)
}

func TestUpdaterConfigError(t *testing.T) {
	testServer := testServerForUpdateFile(t, testZipPath)
	defer testServer.Close()

	upr, _ := newTestUpdaterWithServer(t, testServer, testUpdate(testServer.URL), &testConfig{err: fmt.Errorf("Test config error")})
	ctx := newTestContext(newDefaultTestUpdateOptions(), upr.config, UpdateActionApply)

	_, err := upr.Update(ctx)
	assert.NoError(t, err)

	require.NotNil(t, ctx.errReported)
	assert.Equal(t, ConfigError, ctx.errReported.(Error).errorType)
}

func TestUpdaterAuto(t *testing.T) {
	testServer := testServerForUpdateFile(t, testZipPath)
	defer testServer.Close()

	upr, _ := newTestUpdaterWithServer(t, testServer, testUpdate(testServer.URL), &testConfig{auto: true, autoSet: true})
	ctx := newTestContext(newDefaultTestUpdateOptions(), upr.config, UpdateActionApply)

	_, err := upr.Update(ctx)
	assert.NoError(t, err)
	assert.Equal(t, UpdateActionAuto, ctx.actionReported)
}

func TestUpdaterDownloadNil(t *testing.T) {
	upr, err := newTestUpdater(t)
	require.NoError(t, err)
	err = upr.downloadAsset(nil, UpdateOptions{})
	assert.EqualError(t, err, "No asset to download")
}

func TestUpdaterApplyError(t *testing.T) {
	testServer := testServerForUpdateFile(t, testZipPath)
	defer testServer.Close()

	upr, _ := newTestUpdaterWithServer(t, testServer, testUpdate(testServer.URL), &testConfig{auto: true, autoSet: true})
	ctx := newTestContext(newDefaultTestUpdateOptions(), upr.config, UpdateActionApply)

	ctx.beforeApplyErr = fmt.Errorf("Test before error")
	_, err := upr.Update(ctx)
	assert.EqualError(t, err, "Update Error (apply): Test before error")
	ctx.beforeApplyErr = nil

	ctx.afterApplyErr = fmt.Errorf("Test after error")
	_, err = upr.Update(ctx)
	assert.EqualError(t, err, "Update Error (apply): Test after error")
}

func TestUpdaterNotNeeded(t *testing.T) {
	testServer := testServerForUpdateFile(t, testZipPath)
	defer testServer.Close()

	upr, err := newTestUpdaterWithServer(t, testServer, newTestUpdate(testServer.URL, false), &testConfig{})
	assert.NoError(t, err)
	ctx := newTestContext(newDefaultTestUpdateOptions(), upr.config, UpdateActionSnooze)
	update, err := upr.Update(ctx)
	assert.NoError(t, err)
	assert.Nil(t, update)
}
