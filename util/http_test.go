// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscardAndCloseBodyNil(t *testing.T) {
	err := DiscardAndCloseBody(nil)
	if err == nil {
		t.Fatal("Should have errored")
	}
}

func testServer(t *testing.T, data string, delay time.Duration) *httptest.Server {
	return testServerWithETag(t, data, delay, "")
}

func testServerWithETag(t *testing.T, data string, delay time.Duration, etag string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if delay > 0 {
			time.Sleep(delay)
		}

		etagMatch := r.Header.Get("If-None-Match")
		if etagMatch != "" {
			t.Logf("Checking etag match: %s == %s", etag, etagMatch)
			if etag == etagMatch {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, data)
	}))
}

func testServerForError(err error) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, err.Error(), 500)
	}))
}

func TestSaveHTTPResponse(t *testing.T) {
	data := `{"test": true}`
	server := testServer(t, data, 0)
	defer server.Close()
	resp, err := http.Get(server.URL)
	assert.NoError(t, err)

	savePath := TempPath("", "TestSaveHTTPResponse.")
	defer RemoveFileAtPath(savePath)

	err = SaveHTTPResponse(resp, savePath, 0600, log)
	assert.NoError(t, err)

	saved, err := ioutil.ReadFile(savePath)
	assert.NoError(t, err)

	assert.Equal(t, string(saved), data+"\n")
}

func TestSaveHTTPResponseInvalidPath(t *testing.T) {
	data := `{"test": true}`
	server := testServer(t, data, 0)
	defer server.Close()
	resp, err := http.Get(server.URL)
	assert.NoError(t, err)

	savePath := TempPath("", "TestSaveHTTPResponse.")
	defer RemoveFileAtPath(savePath)

	err = SaveHTTPResponse(resp, "/badpath", 0600, log)
	assert.Error(t, err)
	err = SaveHTTPResponse(nil, savePath, 0600, log)
	assert.Error(t, err)
}

func TestURLExistsValid(t *testing.T) {
	server := testServer(t, "ok", 0)
	defer server.Close()
	exists, err := URLExists(server.URL, time.Second, log)
	assert.True(t, exists)
	assert.NoError(t, err)
}

func TestURLExistsInvalid(t *testing.T) {
	exists, err := URLExists("", time.Second, log)
	assert.Error(t, err)
	assert.False(t, exists)

	exists, err = URLExists("badurl", time.Second, log)
	assert.Error(t, err)
	assert.False(t, exists)

	exists, err = URLExists("http://n", time.Second, log)
	assert.Error(t, err)
	assert.False(t, exists)
}

func TestURLExistsTimeout(t *testing.T) {
	server := testServer(t, "timeout", time.Second)
	defer server.Close()
	exists, err := URLExists(server.URL, time.Millisecond, log)
	t.Logf("Timeout error: %s", err)
	assert.Error(t, err)
	assert.False(t, exists)
}

func TestURLExistsFile(t *testing.T) {
	path, err := WriteTempFile("TestURLExistsFile", []byte(""), 0600)
	assert.NoError(t, err)
	exists, err := URLExists(fmt.Sprintf("file://%s", path), 0, log)
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = URLExists("file:///invalid", 0, log)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestDownloadURLValid(t *testing.T) {
	server := testServer(t, "ok", 0)
	defer server.Close()
	destinationPath := TempPath("", "TestDownloadURLValid.")
	digest, err := Digest(bytes.NewReader([]byte("ok\n")))
	assert.NoError(t, err)
	err = DownloadURL(server.URL, destinationPath, DownloadURLOptions{Digest: digest, RequireDigest: true, Log: log})
	if assert.NoError(t, err) {
		// Check file saved and correct data
		fileExists, fileErr := FileExists(destinationPath)
		assert.NoError(t, fileErr)
		assert.True(t, fileExists)
		data, readErr := ioutil.ReadFile(destinationPath)
		assert.NoError(t, readErr)
		assert.Equal(t, []byte("ok\n"), data)
	}

	// Repeat test, download again, overwriting destination
	server2 := testServer(t, "ok2", 0)
	defer server2.Close()
	digest2, err := Digest(bytes.NewReader([]byte("ok2\n")))
	assert.NoError(t, err)
	err = DownloadURL(server2.URL, destinationPath, DownloadURLOptions{Digest: digest2, RequireDigest: true, Log: log})
	if assert.NoError(t, err) {
		fileExists2, err := FileExists(destinationPath)
		assert.NoError(t, err)
		assert.True(t, fileExists2)
		data2, err := ioutil.ReadFile(destinationPath)
		assert.NoError(t, err)
		assert.Equal(t, []byte("ok2\n"), data2)
	}
}

func TestDownloadURLInvalid(t *testing.T) {
	destinationPath := TempPath("", "TestDownloadURLInvalid.")

	err := DownloadURL("", destinationPath, DownloadURLOptions{Log: log})
	assert.Error(t, err)

	err = DownloadURL("badurl", destinationPath, DownloadURLOptions{Log: log})
	assert.Error(t, err)

	err = DownloadURL("http://", destinationPath, DownloadURLOptions{Log: log})
	assert.Error(t, err)
}

func TestDownloadURLTimeout(t *testing.T) {
	server := testServer(t, "timeout", time.Second)
	defer server.Close()
	destinationPath := TempPath("", "TestDownloadURLInvalid.")
	err := DownloadURL(server.URL, destinationPath, DownloadURLOptions{Timeout: time.Millisecond, Log: log})
	t.Logf("Timeout error: %s", err)
	assert.Error(t, err)
}

func TestDownloadURLParseError(t *testing.T) {
	err := DownloadURL("invalid", "", DownloadURLOptions{})
	assert.Error(t, err)
}

func TestDownloadURLError(t *testing.T) {
	server := testServerForError(fmt.Errorf("Test error"))
	defer server.Close()

	err := DownloadURL(server.URL, "", DownloadURLOptions{})
	assert.EqualError(t, err, "Responded with 500 Internal Server Error")
}

func TestDownloadURLLocal(t *testing.T) {
	var testZipPath = filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/test.zip")
	destinationPath := TempPath("", "TestDownloadURLLocal.")
	defer RemoveFileAtPath(destinationPath)
	err := DownloadURL(fmt.Sprintf("file://%s", testZipPath), destinationPath, DownloadURLOptions{})
	assert.NoError(t, err)

	exists, err := FileExists(destinationPath)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestDownloadURLETag(t *testing.T) {
	data := []byte("ok\n")
	etag := "eff5bc1ef8ec9d03e640fc4370f5eacd"
	server := testServerWithETag(t, "ok", 0, etag)
	defer server.Close()
	destinationPath := TempPath("", "TestDownloadURLETag.")
	err := ioutil.WriteFile(destinationPath, data, 0600)
	require.NoError(t, err)
	digest, err := Digest(bytes.NewReader(data))
	assert.NoError(t, err)
	cached, err := downloadURL(server.URL, destinationPath, DownloadURLOptions{Digest: digest, RequireDigest: true, UseETag: true, Log: log})
	require.NoError(t, err)
	assert.True(t, cached)
}

func TestURLExistsParseError(t *testing.T) {
	exists, err := URLExists("invalid", time.Millisecond, log)
	assert.False(t, exists)
	assert.Error(t, err)
}

func TestURLExistsError(t *testing.T) {
	server := testServerForError(fmt.Errorf("Test error"))
	defer server.Close()

	exists, err := URLExists(server.URL, time.Second, log)
	assert.False(t, exists)
	assert.EqualError(t, err, "Invalid status code (500)")
}

func TestURLValueForBool(t *testing.T) {
	assert.Equal(t, "0", URLValueForBool(false))
	assert.Equal(t, "1", URLValueForBool(true))
}
