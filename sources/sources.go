// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package sources

import keybase1 "github.com/keybase/go-updater/protocol"

// UpdateSource defines where the updater can find updates
type UpdateSource interface {
	// Description is a short description about the update source
	Description() string
	// FindUpdate finds an update given options
	FindUpdate(options keybase1.UpdateOptions) (*keybase1.Update, error)
}
