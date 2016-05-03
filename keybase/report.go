// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/util"
)

// ReportError notifies the API server of a client updater error
func (c context) ReportError(err error, options updater.UpdateOptions) {
	if err := c.reportError(err, options, defaultEndpoints.err, time.Minute); err != nil {
		c.log.Warningf("Error notifying about an error: %s", err)
	}
}

func (c context) reportError(err error, options updater.UpdateOptions, uri string, timeout time.Duration) error {
	var errorType string
	switch uerr := err.(type) {
	case updater.Error:
		errorType = uerr.TypeString()
	default:
		errorType = string(updater.UnknownError)
	}

	data := url.Values{}
	data.Add("install_id", options.InstallID)
	data.Add("version", options.Version)
	data.Add("upd_version", options.UpdaterVersion)
	data.Add("error_type", errorType)
	data.Add("description", err.Error())
	return c.report(data, uri, timeout)
}

// ReportAction notifies the API server of a client updater action
func (c context) ReportAction(action updater.UpdateAction, options updater.UpdateOptions) {
	if err := c.reportAction(action, options, defaultEndpoints.action, time.Minute); err != nil {
		c.log.Warningf("Error notifying about an action (%s): %s", action, err)
	}
}

func (c context) reportAction(action updater.UpdateAction, options updater.UpdateOptions, uri string, timeout time.Duration) error {
	data := url.Values{}
	data.Add("install_id", options.InstallID)
	data.Add("version", options.Version)
	data.Add("upd_version", options.UpdaterVersion)
	data.Add("action", action.String())
	autoUpdate, _ := c.config.GetUpdateAuto()
	data.Add("auto_update", util.URLValueForBool(autoUpdate))
	return c.report(data, uri, timeout)
}

func (c context) report(data url.Values, uri string, timeout time.Duration) error {
	req, err := http.NewRequest("POST", uri, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: timeout}
	c.log.Infof("Reporting error: %s %v", uri, data)
	resp, err := client.Do(req)
	defer util.DiscardAndCloseBodyIgnoreError(resp)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Notify error returned bad HTTP status %v", resp.Status)
	}
	return nil
}
