// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"os"
	"time"

	"github.com/keybase/go-logging"
)

// UpdateChecker runs updates checks every check duration
type UpdateChecker struct {
	updater      *Updater
	ctx          Context
	ticker       *time.Ticker
	log          logging.Logger
	tickDuration time.Duration // tickDuration is the ticker delay
	count        int           // count is number of times we've checked
}

// NewUpdateChecker creates an update checker
func NewUpdateChecker(updater *Updater, ctx Context, log logging.Logger) UpdateChecker {
	return newUpdateChecker(updater, ctx, log, time.Hour)
}

func newUpdateChecker(updater *Updater, ctx Context, log logging.Logger, tickDuration time.Duration) UpdateChecker {
	return UpdateChecker{
		updater:      updater,
		ctx:          ctx,
		log:          log,
		tickDuration: tickDuration,
	}
}

func (u *UpdateChecker) check() error {
	u.count++
	update, err := u.updater.Update(u.ctx)
	if update != nil {
		// If we received an update from the check let's exit, so the watchdog
		// process (e.g. launchd of darwin) can restart us, no matter what, even if
		// there was an error, and even if the update was or wasn't applied.
		// There is no difference between doing another update check in a loop after
		// delay and restarting the service.
		u.log.Info("Exiting for restart")
		os.Exit(0)
	}
	return err
}

// Check checks for an update.
func (u *UpdateChecker) Check() {
	if err := u.check(); err != nil {
		u.log.Errorf("Error in update: %s", err)
	}
}

// Start starts the update checker. Returns false if we are already running.
func (u *UpdateChecker) Start() bool {
	if u.ticker != nil {
		return false
	}
	u.ticker = time.NewTicker(u.tickDuration)
	go func() {
		u.log.Debug("Starting (ticker)")
		for range u.ticker.C {
			u.log.Debug("Checking for update (ticker)")
			u.Check()
		}
	}()
	return true
}

// Stop stops the update checker
func (u *UpdateChecker) Stop() {
	u.ticker.Stop()
	u.ticker = nil
}

// Count is number of times the check has been called
func (u UpdateChecker) Count() int {
	return u.count
}
