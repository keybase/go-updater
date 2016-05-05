// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater/util"
	"github.com/keybase/saltpack"
	"github.com/keybase/saltpack/basic"
)

// SaltpackVerifyDetachedFileAtPath verifies a file
func SaltpackVerifyDetachedFileAtPath(path string, signature string, validKIDs map[string]bool, log logging.Logger) error {
	file, err := os.Open(path)
	defer util.Close(file)
	if err != nil {
		return err
	}
	err = SaltpackVerifyDetached(file, signature, validKIDs, log)
	if err != nil {
		return fmt.Errorf("Error verifying signature: %s", err)
	}
	return nil
}

func checkSender(key saltpack.BasePublicKey, validKIDs map[string]bool, log logging.Logger) error {
	if key == nil {
		return fmt.Errorf("No key")
	}
	kid := key.ToKID()
	if kid == nil {
		return fmt.Errorf("No KID for key")
	}
	skid := hex.EncodeToString(kid)
	log.Infof("Signed by %s", skid)
	if !validKIDs[skid] {
		return fmt.Errorf("Unknown signer KID: %s", skid)
	}
	log.Debugf("Valid KID: %s", skid)
	return nil
}

// SaltpackVerifyDetached verifies a message signature
func SaltpackVerifyDetached(reader io.Reader, signature string, validKIDs map[string]bool, log logging.Logger) error {
	if reader == nil {
		return fmt.Errorf("No reader")
	}
	check := func(key saltpack.BasePublicKey) error {
		return checkSender(key, validKIDs, log)
	}
	return SaltpackVerifyDetachedCheckSender(reader, []byte(signature), check)
}

// SaltpackVerifyDetachedCheckSender verifies a message signature
func SaltpackVerifyDetachedCheckSender(message io.Reader, signature []byte, checkSender func(saltpack.BasePublicKey) error) error {
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
