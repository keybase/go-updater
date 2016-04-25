// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import "fmt"

// UpdateErrorType is a unique short string denoting the error category
type UpdateErrorType string

const (
	// UpdateUnknownError is for if we had an unknown error
	UpdateUnknownError UpdateErrorType = "unknown"
	// UpdateCancelError is for if we canceled
	UpdateCancelError UpdateErrorType = "cancel"
	// UpdatePromptError is an UI prompt error
	UpdatePromptError UpdateErrorType = "prompt"
	// UpdateUnpackError is an error unpacking the asset
	UpdateUnpackError UpdateErrorType = "unpack"
	// UpdateCheckError is an error checking the asset
	UpdateCheckError UpdateErrorType = "check"
	// UpdateApplyError is an error applying the update
	UpdateApplyError UpdateErrorType = "apply"
	// UpdateFindError is an error trying to find the update
	UpdateFindError UpdateErrorType = "find"
	// UpdateDownloadError is an error trying to download the update
	UpdateDownloadError UpdateErrorType = "download"
	// UpdateSaveError is an error trying to save the update
	UpdateSaveError UpdateErrorType = "save"
	// UpdateDigestError is an error with the digest
	UpdateDigestError UpdateErrorType = "digest"
	// UpdateSignatureError is an error verifying signature
	UpdateSignatureError UpdateErrorType = "signature"
)

func (t UpdateErrorType) String() string {
	return string(t)
}

// UpdateError is an update error with a type/category for reporting
type UpdateError struct {
	errorType UpdateErrorType
	source    error
}

// NewUpdateError constructs an UpdateError from a source error
func NewUpdateError(errorType UpdateErrorType, err error) UpdateError {
	return UpdateError{errorType: errorType, source: err}
}

// TypeString returns a unique short string to denote the error type
func (e UpdateError) TypeString() string {
	return e.errorType.String()
}

// Error returns description for an UpdateError
func (e UpdateError) Error() string {
	if e.source == nil {
		return fmt.Sprintf("Update Error (%s)", e.TypeString())
	}
	return fmt.Sprintf("Update Error (%s): %s", e.TypeString(), e.source.Error())
}

func cancelError(message string) UpdateError {
	return NewUpdateError(UpdateCancelError, fmt.Errorf(message))
}
