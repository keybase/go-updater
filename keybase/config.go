// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/keybase/client/go/libkb"
	"github.com/keybase/client/go/logger"
	keybase1 "github.com/keybase/go-updater/protocol"
)

type config struct {
	path            string
	destinationPath string
	pathToKeybase   string
	log             logger.Logger
	store           store
}

type store struct {
	InstallID string `json:"installId"`
	Auto      bool   `json:"auto"`
	AutoSet   bool   `json:"autoSet"`
}

// newConfig returns a valid config, even if it has an error. The caller can
// decide whether this error is fatal or not, or continue with the default
// config.
func newConfig(cfgDir string, destinationPath string, pathToKeybase string, log logger.Logger) (cfg *config, err error) {
	cfg = &config{
		path:            filepath.Join(cfgDir, "updater.json"),
		destinationPath: destinationPath,
		pathToKeybase:   pathToKeybase,
		log:             log,
	}

	if _, serr := os.Stat(cfg.path); os.IsNotExist(serr) {
		log.Warning("%s doesn't exist", cfg.path)
		return
	}

	var file *os.File
	file, err = os.Open(cfg.path)
	if err != nil {
		return
	}
	if file == nil {
		err = fmt.Errorf("No file")
		return
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg.store)
	return
}

func (c config) save() error {
	b, err := json.MarshalIndent(c.store, "", "  ")
	if err != nil {
		return err
	}
	file := libkb.NewFile(c.path, b, 0600)
	err = file.Save()
	if err != nil {
		return err
	}
	return nil
}

func (c config) GetUpdatePreferenceAuto() (bool, bool) {
	return c.store.Auto, c.store.AutoSet
}

func (c *config) SetUpdatePreferenceAuto(b bool) error {
	c.store.Auto = b
	c.store.AutoSet = true
	return c.save()
}

func (c config) GetInstallID() string {
	return c.store.InstallID
}

func (c config) updaterOptions() (options keybase1.UpdateOptions, err error) {
	options = keybase1.UpdateOptions{
		Platform:        runtime.GOOS,
		Arch:            runtime.GOARCH,
		Channel:         "test",
		DestinationPath: c.destinationPath,
		Env:             "prod",
		InstallID:       c.GetInstallID(),
	}

	var version string
	version, err = c.keybaseExecVersion()
	options.Version = version
	return
}

func (c config) keybaseExecVersion() (string, error) {
	output, err := exec.Command(c.pathToKeybase, "version", "-S").Output()
	if err != nil {
		return "", fmt.Errorf("We couldn't figure out version from keybase executable: %s", err)
	}
	ver := strings.TrimSpace(string(output))
	return ver, nil
}
