// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"fmt"
	"io"

	"github.com/keybase/client/go/logger"
	"github.com/keybase/go-updater"
	keybase1 "github.com/keybase/go-updater/protocol"
	"github.com/keybase/saltpack"
)

var validCodeSigningKIDs = []keybase1.KID{
	"01209092ae4e790763dc7343851b977930f35b16cf43ab0ad900a2af3d3ad5cea1a10a", // keybot (device)
	"0120d3458bbecdfc0d0ae39fec05722c6e3e897c169223835977a8aa208dfcd902d30a", // max (device, home)
	"012065ae849d1949a8b0021b165b0edaf722e2a7a9036e07817e056e2d721bddcc0e0a", // max (paper key)
	"01203a5a45c545ef4f661b8b7573711aaecee3fd5717053484a3a3e725cd68abaa5a0a", // chris (device, ccpro)
	"012003d86864fb20e310590042ad3d5492c3f5d06728620175b03c717c211bfaccc20a", // chris (paper key, clay harbor)
}

type context struct {
	updater *updater.Updater
	config  *config
	log     logger.Logger
}

func newContext(upd *updater.Updater, cfg *config, log logger.Logger) *context {
	ctx := context{}
	ctx.updater = upd
	ctx.config = cfg
	ctx.log = log
	return &ctx
}

func (c context) UpdateOptions() (keybase1.UpdateOptions, error) {
	options, err := c.config.updaterOptions()
	if err != nil {
		// Don't error if we have problems determining options
		c.log.Warning("Error trying to determine update options: %s", err)
	}
	return options, nil
}

func (c context) GetUpdateUI() (updater.UpdateUI, error) {
	return c, nil
}

func (c context) GetLog() logger.Logger {
	return c.log
}

func (c context) Verify(reader io.Reader, signature string) error {
	checkSender := func(key saltpack.SigningPublicKey) error {
		kid := updater.SigningPublicKeyToKeybaseKID(key)
		c.log.Info("Signed by %s", kid)
		if kidIsIn(kid, validCodeSigningKIDs) {
			c.log.Debug("Valid KID")
			return nil
		}
		return fmt.Errorf("Unknown signer KID: %s", kid)
	}
	err := updater.SaltpackVerifyDetached(reader, []byte(signature), checkSender)
	if err != nil {
		return fmt.Errorf("Error verifying signature: %s", err)
	}
	return nil
}

func kidIsIn(k keybase1.KID, list []keybase1.KID) bool {
	for _, h := range list {
		if h.Equal(k) {
			return true
		}
	}
	return false
}

// UpdatePrompt is called when the user needs to accept an update
func (c context) UpdatePrompt(update keybase1.Update, options keybase1.UpdatePromptOptions) (keybase1.UpdatePromptResponse, error) {
	return keybase1.UpdatePromptResponse{
		Action:            keybase1.UpdateActionPerformUpdate,
		AlwaysAutoInstall: false,
	}, nil
}
