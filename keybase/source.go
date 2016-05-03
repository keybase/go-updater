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

// UpdateSource finds releases/updates on keybase.io
type UpdateSource struct {
	cfg      config
	log      logging.Logger
	endpoint string
}

// NewUpdateSource contructs an update source for keybase.io
func NewUpdateSource(cfg config, log logging.Logger) UpdateSource {
	return newUpdateSource(cfg, defaultEndpoints.update, log)
}

func newUpdateSource(cfg config, endpoint string, log logging.Logger) UpdateSource {
	return UpdateSource{
		cfg:      cfg,
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
	return k.findUpdate(options, time.Minute)
}

func (k UpdateSource) findUpdate(options updater.UpdateOptions, timeout time.Duration) (*updater.Update, error) {
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

	autoUpdate, _ := k.cfg.GetUpdateAuto()
	urlValues.Add("auto_update", util.URLValueForBool(autoUpdate))

	u.RawQuery = urlValues.Encode()
	urlString := u.String()

	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: timeout}
	k.log.Infof("Request %#v", urlString)
	resp, err := client.Do(req)
	defer util.DiscardAndCloseBodyIgnoreError(resp)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Find update returned bad HTTP status %v", resp.Status)
	}

	var reader io.Reader = resp.Body
	var update updater.Update
	if err = json.NewDecoder(reader).Decode(&update); err != nil {
		return nil, fmt.Errorf("Invalid API response %s", err)
	}

	k.log.Debugf("Received update: %#v", update)
	return &update, nil
}
