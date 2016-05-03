// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"net/http"
	"testing"
	"time"

	"github.com/keybase/go-updater/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPClient(t *testing.T) {
	req, err := http.NewRequest("GET", "https://api.keybase.io/_/api/1.0/user/lookup.json?github=gabriel", nil)
	require.NoError(t, err)
	client, err := httpClient(time.Minute)
	require.NoError(t, err)
	resp, err := client.Do(req)
	defer util.DiscardAndCloseBodyIgnoreError(resp)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
