// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package saltpack

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/keybase/go-logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testLog = &logging.Logger{Module: "test"}

var validCodeSigningKIDs = map[string]bool{
	"9092ae4e790763dc7343851b977930f35b16cf43ab0ad900a2af3d3ad5cea1a1": true, // keybot (device)
}

const message1 = "This is a test message\n"

// This is the output of running:
//   echo "This is a test message" | keybase sign -d
const signature1 = `BEGIN KEYBASE SALTPACK DETACHED SIGNATURE.
  kXR7VktZdyH7rvq v5wcIkPOwDJ1n11 M8RnkLKQGO2f3Bb fzCeMYz4S6oxLAy
  Cco4N255JFQSlh7 IZiojdPCOssX5DX pEcVEdujw3EsDuI FOTpFB77NK4tqLr
  Dgme7xtCaR4QRl2 hchPpr65lKLKSFy YVZcF2xUVN3gjpM vPFUMwg0JTBAG8x
  Z. END KEYBASE SALTPACK DETACHED SIGNATURE.
`

var testZipPath = filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/test.zip")

// keybase sign -d -i test.zip
const testZipSignature = `BEGIN KEYBASE SALTPACK DETACHED SIGNATURE. kXR7VktZdyH7rvq v5wcIkPOwDJ1n11 M8RnkLKQGO2f3Bb fzCeMYz4S6oxLAy Cco4N255JFzv2PX E6WWdobANV4guJI iEE8XJb6uudCX4x QWZfnamVAaZpXuW vdz65rE7oZsLSdW oxMsbBgG9NVpSJy x3CD6LaC9GlZ4IS ofzkHe401mHjr7M M. END KEYBASE SALTPACK DETACHED SIGNATURE.`

func TestVerify(t *testing.T) {
	reader := bytes.NewReader([]byte(message1))
	err := VerifyDetached(reader, signature1, validCodeSigningKIDs, testLog)
	assert.NoError(t, err)
}

func TestVerifyDetachedFileAtPath(t *testing.T) {
	err := VerifyDetachedFileAtPath(testZipPath, testZipSignature, validCodeSigningKIDs, testLog)
	assert.NoError(t, err)
}

func TestVerifyFail(t *testing.T) {
	invalid := bytes.NewReader([]byte("This is a test message changed\n"))
	err := VerifyDetached(invalid, signature1, validCodeSigningKIDs, testLog)
	require.EqualError(t, err, "invalid signature")
}

func TestVerifyFailDetachedFileAtPath(t *testing.T) {
	err := VerifyDetachedFileAtPath(testZipPath, testZipSignature, map[string]bool{}, testLog)
	require.Error(t, err)
}

func TestVerifyNoValidIDs(t *testing.T) {
	reader := bytes.NewReader([]byte(message1))
	err := VerifyDetached(reader, signature1, nil, testLog)
	require.EqualError(t, err, "Unknown signer KID: 9092ae4e790763dc7343851b977930f35b16cf43ab0ad900a2af3d3ad5cea1a1")
}

func TestVerifyBadValidIDs(t *testing.T) {
	var badCodeSigningKIDs = map[string]bool{
		"whatever": true,
	}

	reader := bytes.NewReader([]byte(message1))
	err := VerifyDetached(reader, signature1, badCodeSigningKIDs, testLog)
	require.EqualError(t, err, "Unknown signer KID: 9092ae4e790763dc7343851b977930f35b16cf43ab0ad900a2af3d3ad5cea1a1")
}

func TestVerifyNilInput(t *testing.T) {
	err := VerifyDetached(nil, signature1, validCodeSigningKIDs, testLog)
	require.EqualError(t, err, "No reader")
}

func TestVerifyNoSignature(t *testing.T) {
	reader := bytes.NewReader([]byte(message1))
	err := VerifyDetached(reader, "", validCodeSigningKIDs, testLog)
	require.Equal(t, io.ErrUnexpectedEOF, err)
}

type testSigningKey struct {
	kid []byte
}

func (t testSigningKey) ToKID() []byte {
	return t.kid
}

func (t testSigningKey) Verify(message []byte, signature []byte) error {
	panic("Unsupported")
}

func TestCheckNilSender(t *testing.T) {
	err := checkSender(nil, validCodeSigningKIDs, testLog)
	require.Error(t, err)
}

func TestCheckNoKID(t *testing.T) {
	err := checkSender(testSigningKey{kid: nil}, validCodeSigningKIDs, testLog)
	require.Error(t, err)
}

func TestVerifyNoFile(t *testing.T) {
	err := VerifyDetachedFileAtPath("/invalid", signature1, validCodeSigningKIDs, testLog)
	assert.Error(t, err)
	require.True(t, strings.HasPrefix(err.Error(), "open /invalid: "))
}
