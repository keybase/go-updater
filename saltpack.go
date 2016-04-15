// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/keybase/go-logging"
	"github.com/keybase/saltpack"
	"github.com/keybase/saltpack/basic"
)

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
	kr := basic.NewKeyring()
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
