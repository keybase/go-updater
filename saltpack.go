// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"bytes"
	"fmt"
	"io"

	"github.com/agl/ed25519"
	"github.com/keybase/client/go/libkb"
	keybase1 "github.com/keybase/go-updater/protocol"
	"github.com/keybase/saltpack"
)

type naclSigningKeyPublic [ed25519.PublicKeySize]byte
type naclSignature [ed25519.SignatureSize]byte

const kidNaclEddsa = 0x20

// SaltpackVerifyDetached verifies a message signature
func SaltpackVerifyDetached(message io.Reader, signature []byte, checkSender func(saltpack.SigningPublicKey) error) error {
	sc, _, err := libkb.ClassifyStream(bytes.NewReader(signature))
	if err != nil {
		return err
	}
	if sc.Format != "saltpack" {
		return fmt.Errorf("Wrong format: %s", sc.Format)
	}

	if !sc.Armored {
		return fmt.Errorf("Expected armored message")
	}

	kr := keyring{}
	var skey saltpack.SigningPublicKey
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
func SigningPublicKeyToKeybaseKID(k saltpack.SigningPublicKey) (ret keybase1.KID) {
	if k == nil {
		return ret
	}
	p := k.ToKID()
	return keybase1.KIDFromRawKey(p, kidNaclEddsa)
}
