// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import "time"

// Asset describes a downloadable file
type Asset struct {
	Name      string `codec:"name" json:"name"`
	URL       string `codec:"url" json:"url"`
	Digest    string `codec:"digest" json:"digest"`
	Signature string `codec:"signature" json:"signature"`
	LocalPath string `codec:"localPath" json:"localPath"`
}

// UpdateType is the update type.
// This is an int type for compatibility.
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
	Version     string     `json:"version"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	InstallID   string     `json:"installId"`
	Type        UpdateType `json:"type"`
	PublishedAt Time       `json:"publishedAt"`
	Asset       *Asset     `json:"asset,omitempty"`
}

// UpdateOptions are options used to find an update
type UpdateOptions struct {
	// Version is the current version of the app
	Version string `json:"version"`
	// Platform is the os type (darwin, windows, linux)
	Platform string `json:"platform"`
	// DestinationPath is where to apply the update to)
	DestinationPath string `json:"destinationPath"`
	// URL can override where the updater looks
	URL string `json:"URL"`
	// Channel is an alternative channel to get updates from (test, prerelease)
	Channel string `json:"channel"`
	// Env is an enviroment or run mode (prod, staging, devel)
	Env string `json:"env"`
	// InstallID is an identifier that the client can send with requests
	InstallID string `json:"installId"`
	// Arch is an architecure description (x64, i386, arm)
	Arch string `json:"arch"`
	// Force is whether to apply the update, even if older or same version
	Force bool `json:"force"`
	// SignaturePath is path to signature file to verify against
	SignaturePath string `json:"signaturePath"`
	// OSVersion is the version of the OS
	OSVersion string `json:"osVersion"`
	// UpdaterVersion is the version of the updater service
	UpdaterVersion string `json:"updaterVersion"`
}

// UpdateAction is the update action requested by the user
type UpdateAction string

const (
	// UpdateActionApply means the user accepted and to perform update
	UpdateActionApply UpdateAction = "apply"
	// UpdateActionAuto means that auto update is set and to perform update
	UpdateActionAuto UpdateAction = "auto"
	// UpdateActionSnooze snoozes an update
	UpdateActionSnooze UpdateAction = "snooze"
	// UpdateActionCancel cancels an update
	UpdateActionCancel UpdateAction = "cancel"
	// UpdateActionError means an error occurred
	UpdateActionError UpdateAction = "error"
)

// UpdatePromptOptions are the options for UpdatePrompt
type UpdatePromptOptions struct {
	AutoUpdate bool `json:"autoUpdate"`
}

// UpdatePromptResponse is the result for UpdatePrompt
type UpdatePromptResponse struct {
	Action     UpdateAction `json:"action"`
	AutoUpdate bool         `json:"autoUpdate"`
}

// UpdateUI is a UI interface
type UpdateUI interface {
	// UpdatePrompt prompts for an update
	UpdatePrompt(Update, UpdateOptions, UpdatePromptOptions) (*UpdatePromptResponse, error)
}

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
