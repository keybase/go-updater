// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"fmt"
	"os"
	"time"

	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/command"
	"github.com/keybase/go-updater/saltpack"
)

// validCodeSigningKIDs are the list of valid code signing IDs for saltpack verify
var validCodeSigningKIDs = map[string]bool{
	"9092ae4e790763dc7343851b977930f35b16cf43ab0ad900a2af3d3ad5cea1a1": true, // keybot (device)
	"d3458bbecdfc0d0ae39fec05722c6e3e897c169223835977a8aa208dfcd902d3": true, // max (device, home)
	"65ae849d1949a8b0021b165b0edaf722e2a7a9036e07817e056e2d721bddcc0e": true, // max (paper key, cry glass)
	"3a5a45c545ef4f661b8b7573711aaecee3fd5717053484a3a3e725cd68abaa5a": true, // chris (device, ccpro)
	"03d86864fb20e310590042ad3d5492c3f5d06728620175b03c717c211bfaccc2": true, // chris (paper key, clay harbor)
	"deaa8ae7d06ea9aa49cc678ec49f2b1e1dddb63683e384db539a8649c47925f9": true, // winbot (device, Build)
}

// Log is the logging interface for the keybase package
type Log interface {
	Debug(...interface{})
	Info(...interface{})
	Debugf(s string, args ...interface{})
	Infof(s string, args ...interface{})
	Warningf(s string, args ...interface{})
	Errorf(s string, args ...interface{})
}

// context is an updater.Context implementation
type context struct {
	// config is updater config
	config Config
	// log is the logger
	log Log
}

// endpoints define all the url locations for reporting, etc
type endpoints struct {
	update  string
	action  string
	success string
	err     string
}

var defaultEndpoints = endpoints{
	update:  "https://api.keybase.io/_/api/1.0/pkg/update.json",
	action:  "https://api.keybase.io/_/api/1.0/pkg/act.json",
	success: "https://api.keybase.io/_/api/1.0/pkg/success.json",
	err:     "https://api.keybase.io/_/api/1.0/pkg/error.json",
}

func newContext(cfg Config, log Log) *context {
	ctx := context{
		config: cfg,
		log:    log,
	}
	return &ctx
}

// NewUpdaterContext returns an updater context for Keybase
func NewUpdaterContext(pathToKeybase string, log Log) (updater.Context, *updater.Updater) {
	cfg, err := newConfig("Keybase", pathToKeybase, log)
	if err != nil {
		log.Warningf("Error loading config for context: %s", err)
	}

	src := NewUpdateSource(cfg, log)
	// For testing
	// (cd /Applications; ditto -c -k --sequesterRsrc --keepParent Keybase.app /tmp/Keybase.zip)
	//src := updater.NewLocalUpdateSource("/tmp/Keybase.zip", log)
	upd := updater.NewUpdater(src, &cfg, log)
	return newContext(&cfg, log), upd
}

// UpdateOptions returns update options
func (c *context) UpdateOptions() updater.UpdateOptions {
	return c.config.updaterOptions()
}

// GetUpdateUI returns Update UI
func (c *context) GetUpdateUI() updater.UpdateUI {
	return c
}

// GetLog returns log
func (c context) GetLog() Log {
	return c.log
}

// Verify verifies the signature
func (c context) Verify(update updater.Update) error {
	return saltpack.VerifyDetachedFileAtPath(update.Asset.LocalPath, update.Asset.Signature, validCodeSigningKIDs, c.log)
}

type checkInUseResult struct {
	InUse bool `json:"in_use"`
}

func (c context) checkInUse() (bool, error) {
	var result checkInUseResult
	if err := command.ExecForJSON(c.config.keybasePath(), []string{"update", "check-in-use"}, &result, time.Minute, c.log); err != nil {
		return false, err
	}
	return result.InUse, nil
}

// BeforeApply is called before an update is applied
func (c context) BeforeApply(update updater.Update) error {
	inUse, err := c.checkInUse()
	if err != nil {
		c.log.Warningf("Error trying to check in use: %s", err)
	}
	if inUse {
		if cancel := c.PausedPrompt(); cancel {
			return fmt.Errorf("Canceled by user from paused prompt")
		}
	}
	return nil
}

// AfterApply is called after an update is applied
func (c context) AfterApply(update updater.Update) error {
	result, err := command.Exec(c.config.keybasePath(), []string{"update", "notify", "after-apply"}, 2*time.Minute, c.log)
	if err != nil {
		c.log.Warningf("Error in after apply: %s (%s)", err, result.CombinedOutput())
	}
	return nil
}

func (c context) AfterUpdateCheck(update *updater.Update) {
	if update != nil {
		// If we received an update from the check let's exit, so the watchdog
		// process (e.g. launchd on darwin) can restart us, no matter what, even if
		// there was an error, and even if the update was or wasn't applied.
		// There is no difference between doing another update check in a loop after
		// delay and restarting the service.
		c.log.Infof("%s", "Exiting for restart")
		os.Exit(0)
	}
}
