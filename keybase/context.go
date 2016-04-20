// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"fmt"
	"time"

	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater"
)

// validCodeSigningKIDs are the list of valid code signing IDs for saltpack verify
var validCodeSigningKIDs = map[string]bool{
	"9092ae4e790763dc7343851b977930f35b16cf43ab0ad900a2af3d3ad5cea1a1": true, // keybot (device)
	"d3458bbecdfc0d0ae39fec05722c6e3e897c169223835977a8aa208dfcd902d3": true, // max (device, home)
	"65ae849d1949a8b0021b165b0edaf722e2a7a9036e07817e056e2d721bddcc0e": true, // max (paper key)
	"3a5a45c545ef4f661b8b7573711aaecee3fd5717053484a3a3e725cd68abaa5a": true, // chris (device, ccpro)
	"03d86864fb20e310590042ad3d5492c3f5d06728620175b03c717c211bfaccc2": true, // chris (paper key, clay harbor)
}

// context is an updater.Context implementation
type context struct {
	// updater is the configured updater
	updater *updater.Updater
	// config is updater config
	config *config
	// log is the logger
	log logging.Logger
}

func newContext(upd *updater.Updater, cfg *config, log logging.Logger) *context {
	ctx := context{
		updater: upd,
		config:  cfg,
		log:     log,
	}
	return &ctx
}

// NewUpdaterContext returns an updater context for Keybase
func NewUpdaterContext(pathToKeybase string, log logging.Logger) (updater.Context, *updater.Updater) {
	cfg, err := newConfig("Keybase", pathToKeybase, log)
	if err != nil {
		log.Warningf("Error loading config for context: %s", err)
	}

	src := NewUpdateSource(log)
	// For testing
	//src := sources.NewLocalUpdateSource("/tmp/Keybase.zip", log)
	upd := updater.NewUpdater(src, &cfg, log)
	return newContext(upd, &cfg, log), upd
}

// UpdateOptions returns update options
func (c *context) UpdateOptions() updater.UpdateOptions {
	return c.config.updaterOptions()
}

// GetUpdateUI returns Update UI
func (c *context) GetUpdateUI() (updater.UpdateUI, error) {
	return c, nil
}

// GetLog returns log
func (c context) GetLog() logging.Logger {
	return c.log
}

// Verify verifies the signature
func (c context) Verify(update updater.Update) error {
	return updater.SaltpackVerifyDetachedFileAtPath(update.Asset.LocalPath, update.Asset.Signature, validCodeSigningKIDs, c.log)
}

// BeforeApply is called before an update is applied
func (c context) BeforeApply(update updater.Update) error {
	return nil
}

// AfterApply is called before an update is applied
func (c context) AfterApply(update updater.Update) error {
	output, err := updater.RunCommand(c.config.pathToKeybase, []string{"update", "notify", "after-apply"}, 2*time.Minute, c.log)
	if err != nil {
		return fmt.Errorf("Error in after apply: %s (%s)", err, output)
	}
	return nil
}

func (c context) Restart() error {
	return nil
}
