// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater/util"
)

// RemoteUpdateSource finds releases/updates from custom url feed (used primarily for testing)
type RemoteUpdateSource struct {
	defaultURI string
	log        logging.Logger
}

// NewRemoteUpdateSource builds remote update source without defaults. The url used is passed
// via options instead.
func NewRemoteUpdateSource(defaultURI string, log logging.Logger) RemoteUpdateSource {
	return RemoteUpdateSource{
		defaultURI: defaultURI,
		log:        log,
	}
}

// Description returns update source description
func (r RemoteUpdateSource) Description() string {
	return "Remote"
}

func (r RemoteUpdateSource) sourceURL(options UpdateOptions) string {
	params := util.JoinPredicate([]string{options.Platform, options.Env, options.Channel}, "-", func(s string) bool { return s != "" })
	url := options.URL
	if url == "" {
		url = r.defaultURI
	}
	return fmt.Sprintf("%s/update-%s.json", url, params)
}

// FindUpdate returns update for options
func (r RemoteUpdateSource) FindUpdate(options UpdateOptions) (update *Update, err error) {
	sourceURL := r.sourceURL(options)
	req, err := http.NewRequest("GET", sourceURL, nil)
	client := &http.Client{
		Timeout: time.Minute,
	}
	r.log.Infof("Request %#v", sourceURL)
	resp, err := client.Do(req)
	defer func() { _ = util.DiscardAndCloseBody(resp) }()
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Updater remote returned bad status %v", resp.Status)
		return
	}

	var reader io.Reader = resp.Body
	var obj Update
	if err = json.NewDecoder(reader).Decode(&obj); err != nil {
		err = fmt.Errorf("Bad updater remote response %s", err)
		return
	}
	update = &obj

	r.log.Debugf("Received update %#v", update)

	return
}
