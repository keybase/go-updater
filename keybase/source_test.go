// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keybase/go-updater"
	"github.com/stretchr/testify/assert"
)

const updateJSONResponse = `{
		"version": "1.0.15-20160414190014+fdfce90",
		"name": "v1.0.15-20160414190014+fdfce90",
		"installId": "deadbeef",
		"description": "This is an update!",
		"type": 0,
		"publishedAt": 1460660414000,
		"asset": {
			"name": "Keybase-1.0.15-20160414190014+fdfce90.zip",
			"url": "https://prerelease.keybase.io/darwin-updates/Keybase-1.0.15-20160414190014%2Bfdfce90.zip",
			"digest": "65675b91d0a05f98fcfb44c260f1f6e2c5ba6d6c9d37c84f873c75b65be7d9c4",
			"signature": "BEGIN KEYBASE SALTPACK DETACHED SIGNATURE. kXR7VktZdyH7rvq v5wcIkPOwDJ1n11 M8RnkLKQGO2f3Bb fzCeMYz4S6oxLAy Cco4N255JFgnUxK yZ7SITOx8887cOR aeLbQGWBTMZWEQR hL6bhOCR8CqdXaQ 71lCQkT4WsnqAZe 7bbU2Xrsl50sLbJ BN19a9r6bQBYjce gfK0xY0064VY6CW 9. END KEYBASE SALTPACK DETACHED SIGNATURE.\n",
			"localPath": ""
		}
	}`

func newServer(response string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, response)
	}))
}

func newServerForError(err error) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, err.Error(), 500)
	}))
}

func TestUpdateSource(t *testing.T) {
	server := newServer(updateJSONResponse)
	defer server.Close()

	updateSource := newUpdateSource(server.URL, log)
	options := updater.UpdateOptions{}
	update, err := updateSource.FindUpdate(options)
	assert.NoError(t, err)
	assert.Equal(t, update.Version, "1.0.15-20160414190014+fdfce90")
	assert.Equal(t, update.Name, "v1.0.15-20160414190014+fdfce90")
	assert.Equal(t, update.InstallID, "deadbeef")
	assert.Equal(t, update.Description, "This is an update!")
	assert.True(t, update.PublishedAt == 1460660414000)
	assert.Equal(t, update.Asset.Name, "Keybase-1.0.15-20160414190014+fdfce90.zip")
	assert.Equal(t, update.Asset.URL, "https://prerelease.keybase.io/darwin-updates/Keybase-1.0.15-20160414190014%2Bfdfce90.zip")
}

func TestUpdateSourceBadResponse(t *testing.T) {
	server := newServerForError(fmt.Errorf("Bad response"))
	defer server.Close()

	updateSource := newUpdateSource(server.URL, log)
	options := updater.UpdateOptions{}
	update, err := updateSource.FindUpdate(options)
	t.Logf("Error: %s", err)
	assert.Error(t, err)
	assert.Nil(t, update, "Shouldn't have update")
}
