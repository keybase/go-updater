// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/agl/ed25519"
	"github.com/keybase/go-logging"
	"github.com/keybase/saltpack"
)

type naclSigningKeyPublic [ed25519.PublicKeySize]byte
type naclSignature [ed25519.SignatureSize]byte

const kidNaclEddsa = 0x20

// SaltpackVerifyDetached verifies a message signature
func SaltpackVerifyDetached(reader io.Reader, signature string, validKIDs map[string]bool, log logging.Logger) error {
	if reader == nil {
		return fmt.Errorf("Saltpack Error: No reader")
	}
	checkSender := func(key saltpack.SigningPublicKey) error {
		if key == nil {
			return fmt.Errorf("No key")
		}
		kid := key.ToKID()
		if kid == nil {
			return fmt.Errorf("No KID for key")
		}
		skid := hex.EncodeToString(kid)
		log.Infof("Signed by %s", skid)
		if validKIDs[skid] {
			log.Debug("Valid KID")
			return nil
		}
		return fmt.Errorf("Unknown signer KID: %s", skid)
	}
	return SaltpackVerifyDetachedCheckSender(reader, []byte(signature), checkSender)
}

// SaltpackVerifyDetachedCheckSender verifies a message signature
func SaltpackVerifyDetachedCheckSender(message io.Reader, signature []byte, checkSender func(saltpack.SigningPublicKey) error) error {
	kr := keyring{}
	var skey saltpack.SigningPublicKey
	var err error
	skey, _, err = saltpack.Dearmor62VerifyDetachedReader(message, string(signature), kr)
	if err != nil {
		return err
	}

	if err = checkSender(skey); err != nil {
		return err
	}

	return nil
}

type keyring struct{}

func (e keyring) LookupSigningPublicKey(kid []byte) saltpack.SigningPublicKey {
	var k naclSigningKeyPublic
	copy(k[:], kid)
	return saltSignerPublic{key: k}
}

type saltSignerPublic struct {
	key naclSigningKeyPublic
}

func (s saltSignerPublic) ToKID() []byte {
	return s.key[:]
}

func (s saltSignerPublic) Verify(msg, sig []byte) error {
	if len(sig) != ed25519.SignatureSize {
		return fmt.Errorf("signature size: %d, expected %d", len(sig), ed25519.SignatureSize)
	}

	var fixed naclSignature
	copy(fixed[:], sig)
	if !s.key.Verify(msg, &fixed) {
		return fmt.Errorf("Bad signature")
	}

	return nil
}

func (k naclSigningKeyPublic) Verify(msg []byte, sig *naclSignature) bool {
	return ed25519.Verify((*[ed25519.PublicKeySize]byte)(&k), msg, (*[ed25519.SignatureSize]byte)(sig))
}
