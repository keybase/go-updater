// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater/util"
)

// LocalUpdateSource finds releases/updates from a path (used primarily for testing)
type LocalUpdateSource struct {
	path string
	log  logging.Logger
}

// NewLocalUpdateSource returns local update source
func NewLocalUpdateSource(path string, log logging.Logger) LocalUpdateSource {
	return LocalUpdateSource{
		path: path,
		log:  log,
	}
}

// Description is local update source description
func (k LocalUpdateSource) Description() string {
	return "Local"
}

func digest(URL string) (digest string, err error) {
	f, err := os.Open(URL[7:]) // Remove file:// prefix
	if err != nil {
		return
	}
	defer util.Close(f)
	hasher := sha256.New()
	if _, ioerr := io.Copy(hasher, f); ioerr != nil {
		err = ioerr
		return
	}
	digest = hex.EncodeToString(hasher.Sum(nil))
	return
}

func readFile(path string) (string, error) {
	sigFile, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer util.Close(sigFile)
	data, err := ioutil.ReadAll(sigFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FindUpdate returns update for options
func (k LocalUpdateSource) FindUpdate(options UpdateOptions) (update *Update, err error) {
	url := fmt.Sprintf("file://%s", k.path)

	digest, err := digest(url)
	if err != nil {
		return nil, err
	}
	var signature string
	if options.SignaturePath != "" {
		signature, err = readFile(options.SignaturePath)
		if err != nil {
			return nil, err
		}
	}
	return &Update{
		Version: options.Version,
		Name:    fmt.Sprintf("v%s", options.Version),
		Asset: &Asset{
			Name:      fmt.Sprintf("Keybase-%s.zip", options.Version),
			URL:       url,
			Digest:    digest,
			Signature: signature,
		},
	}, nil
}
