// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/keybase/go-logging"
)

func discardAndClose(rc io.ReadCloser) error {
	_, _ = io.Copy(ioutil.Discard, rc)
	return rc.Close()
}

// DiscardAndCloseBody reads as much as possible from the body of the
// given response, and then closes it.
//
// This is because, in order to free up the current connection for
// re-use, a response body must be read from before being closed; see
// http://stackoverflow.com/a/17953506 .
//
// Instead of doing:
//
//   res, _ := ...
//   defer res.Body.Close()
//
// do
//
//   res, _ := ...
//   defer DiscardAndCloseBody(res)
//
// instead.
func DiscardAndCloseBody(resp *http.Response) error {
	if resp == nil {
		return fmt.Errorf("Nothing to discard (http.Response was nil)")
	}
	return discardAndClose(resp.Body)
}

// SaveHTTPResponse saves an http.Response to path
func SaveHTTPResponse(resp *http.Response, savePath string, mode os.FileMode, log logging.Logger) error {
	if resp == nil {
		return fmt.Errorf("No response")
	}
	file, err := os.OpenFile(savePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	log.Infof("Downloading to %s", savePath)
	n, err := io.Copy(file, resp.Body)
	if err == nil {
		log.Infof("Downloaded %d bytes", n)
	}
	return err
}
