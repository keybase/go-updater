// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build darwin

package keybase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppBundleForPath(t *testing.T) {
	assert.Equal(t, "", appBundleForPath(""))
	assert.Equal(t, "", appBundleForPath("foo"))
	assert.Equal(t, "/Applications/Keybase.app", appBundleForPath("/Applications/Keybase.app"))
	assert.Equal(t, "/Applications/Keybase.app", appBundleForPath("/Applications/Keybase.app/Contents/SharedSupport/bin/keybase"))
	assert.Equal(t, "/Applications/Keybase.app", appBundleForPath("/Applications/Keybase.app/Contents/Resources/Foo.app/Contents/MacOS/Foo"))
	assert.Equal(t, "", appBundleForPath("/Applications/Keybase.ap"))
	assert.Equal(t, "/Applications/Keybase.app", appBundleForPath("/Applications/Keybase.app/"))
}
