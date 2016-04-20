// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// This message signed by keybot
const testMessage = "This is a test message\n"
const testSignature = `BEGIN KEYBASE SALTPACK DETACHED SIGNATURE.
	kXR7VktZdyH7rvq v5wcIkPOwDJ1n11 M8RnkLKQGO2f3Bb fzCeMYz4S6oxLAy
	Cco4N255JFQSlh7 IZiojdPCOssX5DX pEcVEdujw3EsDuI FOTpFB77NK4tqLr
	Dgme7xtCaR4QRl2 hchPpr65lKLKSFy YVZcF2xUVN3gjpM vPFUMwg0JTBAG8x
	Z. END KEYBASE SALTPACK DETACHED SIGNATURE.
`

const testSignatureInvalidSigner = `BEGIN KEYBASE SALTPACK DETACHED SIGNATURE.
	kXR7VktZdyH7rvq v5wcIkPOwGV4GkV Zj40Ut1jYS2euBu Ti6z39EdDX7Ne1P
	i0ToOCpSPXyNeSm Zr6r5UOEZnblXeU gLhEpUSRpLFMlKe MWkq61Yaa8XyFvt
	29NjGzUokNPHPB2 A97cMmFTeGP6Y5V RNRhtwBT3iJoyMv E9RcQhs1717z2aa
	c. END KEYBASE SALTPACK DETACHED SIGNATURE.`

func testContext(t *testing.T) *context {
	cfg, _ := testConfig(t)
	ctx := newContext(&cfg, log)
	assert.NotNil(t, ctx)
	return ctx
}

func TestContext(t *testing.T) {
	ctx := testContext(t)

	// Check options not empty
	options := ctx.UpdateOptions()
	assert.NotEqual(t, options.Version, "")

	reader := bytes.NewReader([]byte(testMessage))
	err := ctx.Verify(reader, testSignature)
	assert.NoError(t, err)
}

func TestContextVerifyFail(t *testing.T) {
	ctx := testContext(t)
	reader := bytes.NewReader([]byte("other data"))
	err := ctx.Verify(reader, testSignature)
	assert.Error(t, err)
}

func TestContextVerifyNoValidIDs(t *testing.T) {
	ctx := testContext(t)
	reader := bytes.NewReader([]byte(testMessage))
	err := ctx.Verify(reader, testSignatureInvalidSigner)
	assert.Error(t, err)
	assert.Equal(t, "Error verifying signature: Unknown signer KID: ad6ec4c0132ca7627b3c4d72c650323abec004da51dc086fd0ec2b4f82e6e486", err.Error())
}

func TestContextVerifyBadSignature(t *testing.T) {
	ctx := testContext(t)
	reader := bytes.NewReader([]byte(testMessage))
	err := ctx.Verify(reader, "BEGIN KEYBASE SALTPACK DETACHED SIGNATURE. END KEYBASE SALTPACK DETACHED SIGNATURE.")
	assert.Error(t, err)
}
