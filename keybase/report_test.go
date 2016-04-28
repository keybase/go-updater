// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/keybase/go-updater"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testOptions = updater.UpdateOptions{
	InstallID: "deadbeef",
	Version:   "1.2.3-400+abcdef",
}

func TestReportError(t *testing.T) {
	server := newServer("{}")
	defer server.Close()

	updateErr := updater.NewError(updater.PromptError, fmt.Errorf("Test error"))
	ctx := testContext(t)
	err := ctx.reportError(updateErr, testOptions, server.URL, 5*time.Millisecond)
	assert.NoError(t, err)
}

func TestReportErrorEmpty(t *testing.T) {
	server := newServer("{}")
	defer server.Close()

	updateErr := updater.NewError(updater.UnknownError, nil)
	emptyOptions := updater.UpdateOptions{}
	ctx := testContext(t)
	err := ctx.reportError(updateErr, emptyOptions, server.URL, 5*time.Millisecond)
	assert.NoError(t, err)
}

func TestReportBadResponse(t *testing.T) {
	server := newServerForError(fmt.Errorf("Bad response"))
	defer server.Close()

	ctx := testContext(t)
	err := ctx.report(url.Values{}, server.URL, 5*time.Millisecond)
	assert.EqualError(t, err, "Notify error returned bad HTTP status 500 Internal Server Error")
}

func TestReportTimeout(t *testing.T) {
	server := newServerWithDelay(updateJSONResponse, 5*time.Millisecond)
	defer server.Close()

	ctx := testContext(t)
	err := ctx.report(url.Values{}, server.URL, 2*time.Millisecond)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "net/http: request canceled"))
}

func TestReportActionApply(t *testing.T) {
	server := newServer("{}")
	defer server.Close()

	ctx := testContext(t)
	err := ctx.reportAction(updater.UpdateActionApply, testOptions, server.URL, 5*time.Millisecond)
	assert.NoError(t, err)
}

func TestReportActionEmpty(t *testing.T) {
	server := newServer("{}")
	defer server.Close()

	ctx := testContext(t)
	err := ctx.reportAction("", updater.UpdateOptions{}, server.URL, 5*time.Millisecond)
	assert.NoError(t, err)
}
