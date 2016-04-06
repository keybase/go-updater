// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package sources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/keybase/client/go/libkb"
	"github.com/keybase/client/go/logger"
	keybase1 "github.com/keybase/go-updater/protocol"
)

// RemoteUpdateSource finds releases/updates from custom url feed (used primarily for testing)
type RemoteUpdateSource struct {
	defaultURI string
	log        logger.Logger
}

// NewRemoteUpdateSource builds remote update source without defaults. The url used is passed
// via options instead.
func NewRemoteUpdateSource(defaultURI string, log logger.Logger) RemoteUpdateSource {
	return RemoteUpdateSource{
		defaultURI: defaultURI,
		log:        log,
	}
}

// Description returns update source description
func (r RemoteUpdateSource) Description() string {
	return "Remote"
}

func (r RemoteUpdateSource) sourceURL(options keybase1.UpdateOptions) string {
	params := libkb.JoinPredicate([]string{options.Platform, options.Env, options.Channel}, "-", func(s string) bool { return s != "" })
	url := options.URL
	if url == "" {
		url = r.defaultURI
	}
	return fmt.Sprintf("%s/update-%s.json", url, params)
}

// FindUpdate returns update for options
func (r RemoteUpdateSource) FindUpdate(options keybase1.UpdateOptions) (update *keybase1.Update, err error) {
	sourceURL := r.sourceURL(options)
	req, err := http.NewRequest("GET", sourceURL, nil)
	client := &http.Client{}
	r.log.Info("Request %#v", sourceURL)
	resp, err := client.Do(req)
	if resp != nil {
		defer libkb.DiscardAndCloseBody(resp)
	}
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Updater remote returned bad status %v", resp.Status)
		return
	}

	var reader io.Reader = resp.Body
	var obj keybase1.Update
	if err = json.NewDecoder(reader).Decode(&obj); err != nil {
		err = fmt.Errorf("Bad updater remote response %s", err)
		return
	}
	update = &obj

	r.log.Debug("Received update %#v", update)

	return
}
