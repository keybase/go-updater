// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

const testZipSignature = `BEGIN KEYBASE SALTPACK DETACHED SIGNATURE.
	kXR7VktZdyH7rvq v5wcIkPOwDJ1n11 M8RnkLKQGO2f3Bb fzCeMYz4S6oxLAy
	Cco4N255JFdUr6I 7XDXrDPRUsFPSRq RK1iOiDNoTlRIXC u4Zi7tLmajqLHUU
	Eo0ng5CsVDR7e4Y DF4S9ioSsqGaQtX euRrI6tMO2EVmEx pQqidbB0aJsKLle
	B. END KEYBASE SALTPACK DETACHED SIGNATURE.`

func TestSaltpackVerify(t *testing.T) {
	reader := bytes.NewReader([]byte(message1))
	err := SaltpackVerifyDetached(reader, signature1, validCodeSigningKIDs, log)
	assert.NoError(t, err)
}

func TestSaltpackVerifyDetachedFileAtPath(t *testing.T) {
	err := SaltpackVerifyDetachedFileAtPath(testZipPath, testZipSignature, validCodeSigningKIDs, log)
	assert.NoError(t, err)
}

func TestSaltpackVerifyFail(t *testing.T) {
	invalid := bytes.NewReader([]byte("This is a test message changed\n"))
	err := SaltpackVerifyDetached(invalid, signature1, validCodeSigningKIDs, log)
	require.Error(t, err, "Should have failed verify")
}

func TestSaltpackVerifyFailDetachedFileAtPath(t *testing.T) {
	err := SaltpackVerifyDetachedFileAtPath(testZipPath, testZipSignature, map[string]bool{}, log)
	require.Error(t, err)
}

func TestSaltpackVerifyNoValidIDs(t *testing.T) {
	reader := bytes.NewReader([]byte(message1))
	err := SaltpackVerifyDetached(reader, signature1, nil, log)
	require.Error(t, err, "Should have failed verify")
	t.Logf("Error: %s", err)
	assert.Equal(t, "Unknown signer KID: 9092ae4e790763dc7343851b977930f35b16cf43ab0ad900a2af3d3ad5cea1a1", err.Error())
}

func TestSaltpackVerifyBadValidIDs(t *testing.T) {
	var badCodeSigningKIDs = map[string]bool{
		"whatever": true,
	}

	reader := bytes.NewReader([]byte(message1))
	err := SaltpackVerifyDetached(reader, signature1, badCodeSigningKIDs, log)
	require.Error(t, err, "Should have failed verify")
	t.Logf("Error: %s", err)
	require.Equal(t, "Unknown signer KID: 9092ae4e790763dc7343851b977930f35b16cf43ab0ad900a2af3d3ad5cea1a1", err.Error())
}

func TestSaltpackVerifyNilInput(t *testing.T) {
	err := SaltpackVerifyDetached(nil, signature1, validCodeSigningKIDs, log)
	require.Error(t, err, "Should have failed verify")
	t.Logf("Error: %s", err)
}

func TestSaltpackVerifyNoSignature(t *testing.T) {
	reader := bytes.NewReader([]byte(message1))
	err := SaltpackVerifyDetached(reader, "", validCodeSigningKIDs, log)
	require.Error(t, err, "Should have failed verify")
	t.Logf("Error: %s", err)
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

func TestSaltpackCheckNilSender(t *testing.T) {
	err := checkSender(nil, validCodeSigningKIDs, log)
	require.Error(t, err)
}

func TestSaltpackCheckNoKID(t *testing.T) {
	err := checkSender(testSigningKey{kid: nil}, validCodeSigningKIDs, log)
	require.Error(t, err)
}

func TestSaltpackVerifyNoFile(t *testing.T) {
	err := SaltpackVerifyDetachedFileAtPath("/invalid", signature1, validCodeSigningKIDs, log)
	require.Error(t, err)
	t.Logf("Error: %s", err)
}
