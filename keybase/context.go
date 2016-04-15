// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"fmt"
	"io"
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
	// config is updater config
	config *config
	// log is the logger
	log logging.Logger
}

func newContext(cfg *config, log logging.Logger) *context {
	ctx := context{}
	ctx.config = cfg
	ctx.log = log
	return &ctx
}

// UpdateOptions returns update options
func (c *context) UpdateOptions() updater.UpdateOptions {
	return c.config.updaterOptions()
}

// GetLog returns log
func (c context) GetLog() logging.Logger {
	return c.log
}

// Verify verifies the signature
func (c context) Verify(reader io.Reader, signature string) error {
	err := updater.SaltpackVerifyDetached(reader, signature, validCodeSigningKIDs, c.log)
	if err != nil {
		return fmt.Errorf("Error verifying signature: %s", err)
	}
	return nil
}

// BeforeApply is called before an update is applied
func (c context) BeforeApply() error {
	return nil
}

// AfterApply is called before an update is applied
func (c context) AfterApply() error {
	output, err := updater.RunCommand(c.config.pathToKeybase, []string{"update", "notify", "after-apply"}, 10*time.Second, c.log)
	if err != nil {
		return fmt.Errorf("Error in after apply: %s (%s)", err, output)
	}
	return nil
}
