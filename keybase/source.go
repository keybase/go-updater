// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/util"
)

type responseStatus struct {
	Code   int               `json:"code"`
	Name   string            `json:"name"`
	Desc   string            `json:"desc"`
	Fields map[string]string `json:"fields"`
}

type updateResponse struct {
	Status responseStatus `json:"status"`
	Update updater.Update `json:"update"`
}

// UpdateSource finds releases/updates on keybase.io
type UpdateSource struct {
	log      logging.Logger
	endpoint string
}

// NewUpdateSource contructs an update source for keybase.io
func NewUpdateSource(log logging.Logger) UpdateSource {
	return newUpdateSource("https://keybase.io/_/api/1.0/pkg/update.json", log)
}

func newUpdateSource(endpoint string, log logging.Logger) UpdateSource {
	return UpdateSource{
		endpoint: endpoint,
		log:      log,
	}
}

// Description returns description for update source
func (k UpdateSource) Description() string {
	return "Keybase.io"
}

// FindUpdate returns update for updater and options
func (k UpdateSource) FindUpdate(options updater.UpdateOptions) (*updater.Update, error) {
	if options.URL != "" {
		return nil, fmt.Errorf("Custom URLs not supported for this update source")
	}

	u, err := url.Parse(k.endpoint)
	if err != nil {
		return nil, err
	}
	urlValues := url.Values{}
	urlValues.Add("install_id", options.InstallID)
	urlValues.Add("version", options.Version)
	urlValues.Add("platform", options.Platform)
	urlValues.Add("run_mode", options.Env)
	urlValues.Add("os_version", options.OSVersion)
	urlValues.Add("upd_version", options.UpdaterVersion)
	u.RawQuery = urlValues.Encode()
	urlString := u.String()

	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: time.Minute,
	}
	k.log.Infof("Request %#v", urlString)
	resp, err := client.Do(req)
	defer func() { _ = util.DiscardAndCloseBody(resp) }()
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Keybase returned bad HTTP status %v", resp.Status)
	}

	var reader io.Reader = resp.Body
	var res updateResponse
	if err = json.NewDecoder(reader).Decode(&res); err != nil {
		return nil, fmt.Errorf("Invalid API response %s", err)
	}

	if res.Status.Code != 0 {
		return nil, fmt.Errorf("API returned error response: %#v", res)
	}

	k.log.Debugf("Received update: %#v", res.Update)
	return &res.Update, nil
}
