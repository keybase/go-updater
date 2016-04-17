// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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

// DiscardAndCloseBodyIgnoreError calls DiscardAndCloseBody.
// This satisfies lint checks when using with defer and you don't care if there
// is an error, so instead of:
//   defer func() { _ = DiscardAndCloseBody(resp) }()
//   defer DiscardAndCloseBodyIgnoreError(resp)
func DiscardAndCloseBodyIgnoreError(resp *http.Response) {
	_ = DiscardAndCloseBody(resp)
}
