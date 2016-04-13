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

// KID is a key identifier
type KID string

const (
	// KIDLen is KID length in bytes
	KIDLen = 35
	// KIDSuffix is KID suffix (byte)
	KIDSuffix = 0x0a
	// KIDVersion is current version of KID
	KIDVersion = 0x1
)

// KIDFromRawKey returns KID from bytes by type
func KIDFromRawKey(b []byte, keyType byte) KID {
	tmp := []byte{KIDVersion, keyType}
	tmp = append(tmp, b...)
	tmp = append(tmp, byte(KIDSuffix))
	return KIDFromSlice(tmp)
}

// KIDFromSlice returns KID from bytes
func KIDFromSlice(b []byte) KID {
	return KID(hex.EncodeToString(b))
}

// Equal returns true if KID's are equal
func (k KID) Equal(v KID) bool {
	return k == v
}

// SaltpackVerifyDetached verifies a message signature
func SaltpackVerifyDetached(reader io.Reader, signature string, validKIDs []KID, log logging.Logger) error {
	if reader == nil {
		return fmt.Errorf("Saltpack Error: No reader")
	}
	checkSender := func(key saltpack.SigningPublicKey) error {
		kid := SigningPublicKeyToKeybaseKID(key)
		log.Infof("Signed by %s", kid)
		if kidIsIn(kid, validKIDs) {
			log.Debug("Valid KID")
			return nil
		}
		return fmt.Errorf("Unknown signer KID: %s", kid)
	}
	return SaltpackVerifyDetachedCheckSender(reader, []byte(signature), checkSender)
}

func kidIsIn(k KID, list []KID) bool {
	for _, h := range list {
		if h.Equal(k) {
			return true
		}
	}
	return false
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

// SigningPublicKeyToKeybaseKID returns a KID for a public key
func SigningPublicKeyToKeybaseKID(k saltpack.SigningPublicKey) (ret KID) {
	if k == nil {
		return ret
	}
	p := k.ToKID()
	return KIDFromRawKey(p, kidNaclEddsa)
}
