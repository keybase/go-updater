// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"bytes"
	"testing"
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

func TestSaltpackVerify(t *testing.T) {
	reader := bytes.NewReader([]byte(message1))
	err := SaltpackVerifyDetached(reader, signature1, validCodeSigningKIDs, log)
	if err != nil {
		t.Errorf("Error in verify: %s", err)
	}
}

func TestSaltpackVerifyFail(t *testing.T) {
	invalid := bytes.NewReader([]byte("This is a test message changed\n"))
	err := SaltpackVerifyDetached(invalid, signature1, validCodeSigningKIDs, log)
	if err == nil {
		t.Fatal("Should have failed verification")
	}
}

func TestSaltpackVerifyNoValidIDs(t *testing.T) {
	reader := bytes.NewReader([]byte(message1))
	err := SaltpackVerifyDetached(reader, signature1, nil, log)
	t.Logf("Error: %s", err)
	if err == nil {
		t.Fatal("Should have errored")
	}
	if err.Error() != "Unknown signer KID: 9092ae4e790763dc7343851b977930f35b16cf43ab0ad900a2af3d3ad5cea1a1" {
		t.Errorf("Unexpected error output")
	}
}

func TestSaltpackVerifyBadValidIDs(t *testing.T) {
	var badCodeSigningKIDs = map[string]bool{
		"whatever": true,
	}

	reader := bytes.NewReader([]byte(message1))
	err := SaltpackVerifyDetached(reader, signature1, badCodeSigningKIDs, log)
	t.Logf("Error: %s", err)
	if err == nil {
		t.Fatal("Should have errored")
	}
	if err.Error() != "Unknown signer KID: 9092ae4e790763dc7343851b977930f35b16cf43ab0ad900a2af3d3ad5cea1a1" {
		t.Errorf("Unexpected error output")
	}
}

func TestSaltpackVerifyNilInput(t *testing.T) {
	err := SaltpackVerifyDetached(nil, signature1, validCodeSigningKIDs, log)
	t.Logf("Error: %s", err)
	if err == nil {
		t.Fatal("Should have errored")
	}
}

func TestSaltpackVerifyNoSignature(t *testing.T) {
	reader := bytes.NewReader([]byte(message1))
	err := SaltpackVerifyDetached(reader, "", validCodeSigningKIDs, log)
	t.Logf("Error: %s", err)
	if err == nil {
		t.Fatal("Should have errored")
	}
}
