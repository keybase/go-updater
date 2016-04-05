// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package protocol

import (
	"encoding/hex"
	"time"
)

// Time is milliseconds since epoch
type Time int64

// FromTime converts protocol time to golang Time
func FromTime(t Time) time.Time {
	if t == 0 {
		return time.Time{}
	}
	return time.Unix(0, int64(t)*1000000)
}

// ToTime converts golang Time to protocol Time
func ToTime(t time.Time) Time {
	// the result of calling UnixNano on the zero Time is undefined.
	// https://golang.org/pkg/time/#Time.UnixNano
	if t.IsZero() {
		return 0
	}
	return Time(t.UnixNano() / 1000000)
}

// Asset describes a downloadable file
type Asset struct {
	Name      string `codec:"name" json:"name"`
	URL       string `codec:"url" json:"url"`
	Digest    string `codec:"digest" json:"digest"`
	Signature string `codec:"signature" json:"signature"`
	LocalPath string `codec:"localPath" json:"localPath"`
}

// UpdateType is the update type
type UpdateType int

const (
	// UpdateTypeNormal is a normal update
	UpdateTypeNormal UpdateType = 0
	// UpdateTypeBugFix is a bugfix update
	UpdateTypeBugFix UpdateType = 1
	// UpdateTypeCritical is a critical update
	UpdateTypeCritical UpdateType = 2
)

// Update defines an update to apply
type Update struct {
	Version      string     `codec:"version" json:"version"`
	Name         string     `codec:"name" json:"name"`
	Description  string     `codec:"description" json:"description"`
	Instructions string     `codec:"instructions" json:"instructions"`
	Type         UpdateType `codec:"type" json:"type"`
	PublishedAt  Time       `codec:"publishedAt" json:"publishedAt"`
	Asset        *Asset     `codec:"asset,omitempty" json:"asset,omitempty"`
}

// UpdateOptions are options used to find an update
type UpdateOptions struct {
	// Version is the current version of the app
	Version string `codec:"version" json:"version"`
	// Platform is the os type (darwin, windows, linux)
	Platform string `codec:"platform" json:"platform"`
	// DestinationPath is where to apply the update to)
	DestinationPath string `codec:"destinationPath" json:"destinationPath"`
	// Source is the updater source type (keybase, s3)
	Source string `codec:"source" json:"source"`
	// URL can override where the updater looks
	URL string `codec:"URL" json:"URL"`
	// Channel is an alternative channel to get updates from (test, prerelease)
	Channel string `codec:"channel" json:"channel"`
	// Env is an enviroment or run mode (prod, staging, devel)
	Env string `codec:"env" json:"env"`
	// InstallID is an identifier that the client can send with requests
	InstallID string `codec:"installId" json:"installId"`
	// Arch is an architecure description (x64, i386, arm)
	Arch string `codec:"arch" json:"arch"`
	// Force is whether to apply the update, even if older or same version
	Force bool `codec:"force" json:"force"`
	// DefaultInstructions is what update instructions to show
	DefaultInstructions string `codec:"defaultInstructions" json:"defaultInstructions"`
	// SignaturePath is path to signature file to verify against
	SignaturePath string `codec:"signaturePath" json:"signaturePath"`
}

// UpdateAction is the update action requested by the user
type UpdateAction string

const (
	// UpdateActionPerformUpdate performs an update
	UpdateActionPerformUpdate UpdateAction = "perform"
	// UpdateActionSnooze snoozes an update
	UpdateActionSnooze UpdateAction = "snooze"
	// UpdateActionCancel cancels an update
	UpdateActionCancel UpdateAction = "cancel"
)

// UpdatePromptOptions are the options for UpdatePrompt
type UpdatePromptOptions struct {
	AlwaysAutoInstall bool `codec:"alwaysAutoInstall" json:"alwaysAutoInstall"`
}

// UpdatePromptResponse is the result for UpdatePrompt
type UpdatePromptResponse struct {
	Action            UpdateAction `codec:"action" json:"action"`
	AlwaysAutoInstall bool         `codec:"alwaysAutoInstall" json:"alwaysAutoInstall"`
}

// UpdateUI is interface for UI
type UpdateUI interface {
	// UpdatePrompt prompts for an update
	UpdatePrompt(update Update, options UpdatePromptOptions) (UpdatePromptResponse, error)
}

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
