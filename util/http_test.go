// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package util

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscardAndCloseBodyNil(t *testing.T) {
	err := DiscardAndCloseBody(nil)
	if err == nil {
		t.Fatal("Should have errored")
	}
}

func testResponse(t *testing.T, data string) *http.Response {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, data)
	}))

	resp, err := http.Get(server.URL)
	assert.NoError(t, err)
	return resp
}

func TestSaveHTTPResponse(t *testing.T) {
	data := `{"test": true}`
	resp := testResponse(t, data)
	savePath, err := TempPath("TestSaveHTTPResponse.")
	assert.NoError(t, err)
	defer RemoveFileAtPath(savePath)
	err = SaveHTTPResponse(resp, savePath, 0600, log)
	assert.NoError(t, err)

	saved, err := ioutil.ReadFile(savePath)
	assert.NoError(t, err)

	assert.Equal(t, string(saved), data+"\n")
}

func TestSaveHTTPResponseInvalidPath(t *testing.T) {
	data := `{"test": true}`
	resp := testResponse(t, data)
	savePath, err := TempPath("TestSaveHTTPResponse.")
	assert.NoError(t, err)
	defer RemoveFileAtPath(savePath)
	err = SaveHTTPResponse(resp, "/badpath", 0600, log)
	assert.Error(t, err)
	err = SaveHTTPResponse(nil, savePath, 0600, log)
	assert.Error(t, err)
}
