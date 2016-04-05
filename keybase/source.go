// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/keybase/client/go/libkb"
	"github.com/keybase/client/go/logger"
	"github.com/keybase/go-updater"
	keybase1 "github.com/keybase/go-updater/protocol"
)

type responseStatus struct {
	Code   int               `json:"code"`
	Name   string            `json:"name"`
	Desc   string            `json:"desc"`
	Fields map[string]string `json:"fields"`
}

type updateResponse struct {
	Status responseStatus  `json:"status"`
	Update keybase1.Update `json:"update"`
}

// KeybaseUpdateSource finds releases/updates on keybase.io
type KeybaseUpdateSource struct {
	runMode string
	log     logger.Logger
}

// NewKeybaseUpdateSource contructs an update source for keybase.io
func NewKeybaseUpdateSource(log logger.Logger) KeybaseUpdateSource {
	return KeybaseUpdateSource{log: log}
}

// Description returns description for update source
func (k KeybaseUpdateSource) Description() string {
	return "Keybase.io"
}

// FindUpdate returns update for updater and options
func (k KeybaseUpdateSource) FindUpdate(options keybase1.UpdateOptions) (update *keybase1.Update, err error) {
	if options.URL != "" {
		return nil, fmt.Errorf("Custom URLs not supported for this update source")
	}

	u, err := url.Parse("https://keybase.io/_/api/1.0/pkg/update.json")
	urlValues := url.Values{}
	urlValues.Add("install_id", options.InstallID)
	urlValues.Add("version", options.Version)
	urlValues.Add("platform", options.Platform)
	urlValues.Add("channel", options.Channel)
	urlValues.Add("run_mode", options.Env)
	urlValues.Add("upd_version", updater.Version)
	urlValues.Add("os_version", osVersion())
	// TODO: OS version
	// urlValues.Add("os_version", )
	u.RawQuery = urlValues.Encode()
	urlString := u.String()

	req, err := http.NewRequest("GET", urlString, nil)
	client := &http.Client{}
	k.log.Info("Request %#v", urlString)
	resp, err := client.Do(req)
	if resp != nil {
		defer libkb.DiscardAndCloseBody(resp)
	}
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Keybase returned bad HTTP status %v", resp.Status)
		return
	}

	var reader io.Reader = resp.Body
	var res updateResponse
	if err = json.NewDecoder(reader).Decode(&res); err != nil {
		err = fmt.Errorf("Invalid API response %s", err)
		return
	}

	if res.Status.Code != 0 {
		err = fmt.Errorf("API returned error response: %#v", res)
	} else {
		k.log.Debug("Received update: %#v", res.Update)
		update = &res.Update
	}
	return
}
