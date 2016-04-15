// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	cfg, _ := testConfig(t)
	ctx := newContext(&cfg, log)
	assert.NotNil(t, ctx)

	// Check options not empty
	options := ctx.UpdateOptions()
	assert.NotEqual(t, options.Version, "")

	// This message signed by keybot
	const message1 = "This is a test message\n"
	const signature1 = `BEGIN KEYBASE SALTPACK DETACHED SIGNATURE.
    kXR7VktZdyH7rvq v5wcIkPOwDJ1n11 M8RnkLKQGO2f3Bb fzCeMYz4S6oxLAy
    Cco4N255JFQSlh7 IZiojdPCOssX5DX pEcVEdujw3EsDuI FOTpFB77NK4tqLr
    Dgme7xtCaR4QRl2 hchPpr65lKLKSFy YVZcF2xUVN3gjpM vPFUMwg0JTBAG8x
    Z. END KEYBASE SALTPACK DETACHED SIGNATURE.
  `
	reader := bytes.NewReader([]byte(message1))
	err := ctx.Verify(reader, signature1)
	assert.Nil(t, err, "%s", err)
}
