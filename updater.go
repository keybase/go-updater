// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import "github.com/keybase/go-logging"

// Version is the updater version
const Version = "0.2.1"

// Updater knows how to find and apply updates
type Updater struct {
	source UpdateSource
	config Config
	log    logging.Logger
}

// UpdateSource defines where the updater can find updates
type UpdateSource interface {
	// Description is a short description about the update source
	Description() string
	// FindUpdate finds an update given options
	FindUpdate(options UpdateOptions) (*Update, error)
}

// Context defines state during an update session
type Context interface {
	GetUpdateUI() (UpdateUI, error)
	UpdateOptions() UpdateOptions
	Verify(update Update) error
	BeforeApply(update Update) error
	AfterApply(update Update) error
	Restart() error
}

// Config defines configuration for the Updater
type Config interface {
	GetUpdateAuto() (bool, bool)
	SetUpdateAuto(b bool) error
	GetInstallID() string
	SetInstallID(installID string) error
}

// NewUpdater constructs an Updater
func NewUpdater(source UpdateSource, config Config, log logging.Logger) *Updater {
	return &Updater{
		source: source,
		config: config,
		log:    log,
	}
}

// Update performs an update
func (u *Updater) Update(ctx Context) (*Update, error) {
	// TODO
	return nil, nil
}
