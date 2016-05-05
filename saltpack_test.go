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

// keybase sign -d -i test.zip
const testZipSignature = `BEGIN KEYBASE SALTPACK DETACHED SIGNATURE.
	kXR7VktZdyH7rvq v5wcIkPOwDJ1n11 M8RnkLKQGO2f3Bb fzCeMYz4S6oxLAy
	Cco4N255JFTAn8O 78IT0oJCfKGRxGx NGkFZsPsFKcFSjt pXUAgKgYFpxs8XM
	2Nbn5qzg3t3rky3 bX8iMOXbqWLewah a7GnOT5bOlbzf8V 1uhiECJ0N6IvRBp
	D. END KEYBASE SALTPACK DETACHED SIGNATURE.`

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
	require.EqualError(t, err, "invalid signature")
}

func TestSaltpackVerifyFailDetachedFileAtPath(t *testing.T) {
	err := SaltpackVerifyDetachedFileAtPath(testZipPath, testZipSignature, map[string]bool{}, log)
	require.Error(t, err)
}

func TestSaltpackVerifyNoValidIDs(t *testing.T) {
	reader := bytes.NewReader([]byte(message1))
	err := SaltpackVerifyDetached(reader, signature1, nil, log)
	require.EqualError(t, err, "Unknown signer KID: 9092ae4e790763dc7343851b977930f35b16cf43ab0ad900a2af3d3ad5cea1a1")
}

func TestSaltpackVerifyBadValidIDs(t *testing.T) {
	var badCodeSigningKIDs = map[string]bool{
		"whatever": true,
	}

	reader := bytes.NewReader([]byte(message1))
	err := SaltpackVerifyDetached(reader, signature1, badCodeSigningKIDs, log)
	require.EqualError(t, err, "Unknown signer KID: 9092ae4e790763dc7343851b977930f35b16cf43ab0ad900a2af3d3ad5cea1a1")
}

func TestSaltpackVerifyNilInput(t *testing.T) {
	err := SaltpackVerifyDetached(nil, signature1, validCodeSigningKIDs, log)
	require.EqualError(t, err, "No reader")
}

func TestSaltpackVerifyNoSignature(t *testing.T) {
	reader := bytes.NewReader([]byte(message1))
	err := SaltpackVerifyDetached(reader, "", validCodeSigningKIDs, log)
	require.EqualError(t, err, "Error in framing: wrong number of words (1)")
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
	require.EqualError(t, err, "open /invalid: no such file or directory")
}
