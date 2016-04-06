// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/keybase/client/go/libkb"
	"github.com/keybase/client/go/logger"
	keybase1 "github.com/keybase/go-updater/protocol"
)

type processUpdate func(update *keybase1.Update, path string)

func newTestUpdater(t *testing.T, p processUpdate) (*Updater, error) {
	updateSource, err := newTestUpdateSource(p)
	if err != nil {
		return nil, err
	}
	config := &testConfig{}
	return NewUpdater(updateSource, config, logger.NewTestLogger(t)), nil
}

type testUpdateUI struct {
	options keybase1.UpdateOptions
}

func (u testUpdateUI) UpdatePrompt(_ keybase1.Update, _ keybase1.UpdatePromptOptions) (keybase1.UpdatePromptResponse, error) {
	return keybase1.UpdatePromptResponse{Action: keybase1.UpdateActionPerformUpdate}, nil
}

func (u testUpdateUI) GetUpdateUI() (UpdateUI, error) {
	return u, nil
}

func (u testUpdateUI) Verify(r io.Reader, signature string) error {
	digest, err := libkb.Digest(r)
	if err != nil {
		return err
	}
	if signature != digest {
		return fmt.Errorf("Verify failed")
	}
	return nil
}

func (u testUpdateUI) UpdateOptions() (keybase1.UpdateOptions, error) {
	return u.options, nil
}

type testUpdateSource struct {
	processUpdate processUpdate
}

func newTestUpdateSource(p processUpdate) (testUpdateSource, error) {
	return testUpdateSource{processUpdate: p}, nil
}

func (u testUpdateSource) Description() string {
	return "Test"
}

func (u testUpdateSource) FindUpdate(config keybase1.UpdateOptions) (*keybase1.Update, error) {
	version := "1.0.1"
	update := keybase1.Update{
		Version:     version,
		Name:        "Test",
		Description: "Bug fixes",
	}

	path := filepath.Join(os.TempDir(), "Test.zip")
	assetName, err := createTestUpdateFile(path, version)
	if err != nil {
		return nil, err
	}

	if path != "" {
		digest, err := libkb.DigestForFileAtPath(path)
		if err != nil {
			return nil, err
		}

		update.Asset = &keybase1.Asset{
			Name:      assetName,
			URL:       fmt.Sprintf("file://%s", path),
			Digest:    digest,
			Signature: digest, // Use digest as signature in test
		}
	}

	if u.processUpdate != nil {
		u.processUpdate(&update, path)
	}

	return &update, nil
}

type testConfig struct {
	lastChecked  keybase1.Time
	publicKeyHex string
}

func (c testConfig) GetUpdatePreferenceAuto() (bool, bool) {
	return false, false
}

func (c *testConfig) SetUpdatePreferenceAuto(b bool) error {
	return nil
}

func newDefaultTestUpdateOptions() keybase1.UpdateOptions {
	return keybase1.UpdateOptions{
		Version:             "1.0.0",
		Platform:            runtime.GOOS,
		DestinationPath:     filepath.Join(os.TempDir(), "Test"),
		Source:              "test",
		DefaultInstructions: "Bug fixes",
	}
}

func TestUpdater(t *testing.T) {
	u, err := newTestUpdater(t, nil)
	if err != nil {
		t.Fatal(err)
	}
	update, err := u.Update(testUpdateUI{newDefaultTestUpdateOptions()})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Update: %#v\n", update)

	if update.Asset == nil {
		t.Errorf("No asset")
	}

	t.Logf("Asset: %#v\n", *update.Asset)

	if update.Asset.Signature == "" {
		t.Errorf("No signature")
	}

	if update == nil {
		t.Errorf("Should have an update")
	}
}

func TestUpdateCheckErrorIfLowerVersion(t *testing.T) {
	u, err := newTestUpdater(t, nil)
	if err != nil {
		t.Fatal(err)
	}
	options := newDefaultTestUpdateOptions()
	options.Version = "100000000.0.0"

	update, err := u.checkForUpdate(options, true)
	if err != nil {
		t.Fatal(err)
	}
	if update != nil {
		t.Fatal("Shouldn't have update since our version is newer")
	}
}

func TestChangeUpdateFailSignature(t *testing.T) {
	changeAsset := func(u *keybase1.Update, path string) {
		// Write new file over existing (fix digest but not signature)
		_, err := createTestUpdateFile(path, u.Version)
		if err != nil {
			t.Fatal(err)
		}
		digest, _ := libkb.DigestForFileAtPath(path)
		t.Logf("Wrote a new update file: %s (%s)", path, digest)
		u.Asset.Digest = digest
	}
	updater, err := newTestUpdater(t, changeAsset)
	if err != nil {
		t.Fatal(err)
	}
	_, err = updater.Update(testUpdateUI{newDefaultTestUpdateOptions()})
	t.Logf("Err: %s\n", err)
	if err == nil {
		t.Fatal("Should have failed")
	}
}

func randString(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(fmt.Sprintf("Read errored: %s", err))
	}
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}
