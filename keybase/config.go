// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/util"
)

type config struct {
	// appName is the name of the app, e.g. "Keybase"
	appName string
	// pathToKeybase is the location of the keybase executable
	pathToKeybase string
	// log is the logging location
	log logging.Logger
	// store is the config values
	store store
}

// store is the config values
type store struct {
	InstallID string `json:"installId"`
	Auto      bool   `json:"auto"`
	AutoSet   bool   `json:"autoSet"`
}

// newConfig loads a config, which is valid even if it has an error
func newConfig(appName string, pathToKeybase string, log logging.Logger) (config, error) {
	cfg := newDefaultConfig(appName, pathToKeybase, log)
	err := cfg.load()
	return cfg, err
}

func newDefaultConfig(appName string, pathToKeybase string, log logging.Logger) config {
	return config{
		appName:       appName,
		pathToKeybase: pathToKeybase,
		log:           log,
	}
}

// load the config
func (c *config) load() error {
	path, err := c.path()
	if err != nil {
		return nil
	}
	return c.loadFromPath(path)
}

func (c *config) loadFromPath(path string) error {
	if _, serr := os.Stat(path); os.IsNotExist(serr) {
		c.log.Warningf("Unable to load config, %s doesn't exist", path)
		return serr
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	if file == nil {
		return fmt.Errorf("No file")
	}
	defer util.Close(file)

	decoder := json.NewDecoder(file)
	// Instead of seprate store var here, we could decode directly into c.store,
	// but we want to only mutate the store if decoding succeeds completely.
	var decodeStore store
	if err := decoder.Decode(&decodeStore); err != nil {
		return err
	}
	c.store = decodeStore
	return nil
}

func (c config) path() (string, error) {
	configDir, err := c.dir()
	if err != nil {
		return "", err
	}
	if configDir == "" {
		return "", fmt.Errorf("No config dir")
	}
	path := filepath.Join(configDir, "updater.json")
	return path, nil
}

func (c config) save() error {
	path, err := c.path()
	if err != nil {
		return err
	}
	return c.saveToPath(path)
}

func (c config) saveToPath(path string) error {
	b, err := json.MarshalIndent(c.store, "", "  ")
	if err != nil {
		return fmt.Errorf("Error marshaling config: %s", err)
	}
	file := util.NewFile(path, b, 0600)
	err = util.MakeParentDirs(path)
	if err != nil {
		return err
	}
	return file.Save(c.log)
}

func (c config) GetUpdateAuto() (bool, bool) {
	return c.store.Auto, c.store.AutoSet
}

func (c *config) SetUpdateAuto(auto bool) error {
	c.store.Auto = auto
	c.store.AutoSet = true
	return c.save()
}

func (c config) GetInstallID() string {
	return c.store.InstallID
}

func (c *config) SetInstallID(installID string) error {
	c.store.InstallID = installID
	return c.save()
}

func (c config) updaterOptions() updater.UpdateOptions {
	version := c.keybaseExecVersion()
	osVersion := c.osVersion()

	return updater.UpdateOptions{
		Version:         version,
		Platform:        runtime.GOOS,
		Arch:            runtime.GOARCH,
		Channel:         "test",
		DestinationPath: c.destinationPath(),
		Env:             "prod",
		InstallID:       c.GetInstallID(),
		OSVersion:       osVersion,
		UpdaterVersion:  updater.Version,
	}
}

func (c config) keybaseExecVersion() string {
	output, err := updater.RunCommand(c.pathToKeybase, []string{"version", "-S"}, 5*time.Second, c.log)
	if err != nil {
		c.log.Warningf("Couldn't get keybase version: %s", err)
		return ""
	}
	return strings.TrimSpace(output)
}
