// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import "fmt"

// ErrorType is a unique short string denoting the error category
type ErrorType string

const (
	// UnknownError is for if we had an unknown error
	UnknownError ErrorType = "unknown"
	// CancelError is for if we canceled
	CancelError ErrorType = "cancel"
	// ConfigError is for errors reading/saving config
	ConfigError ErrorType = "config"
)

// Errors corresponding to each stage in the update process
const (
	// FindError is an error trying to find the update
	FindError ErrorType = "find"
	// PromptError is an UI prompt error
	PromptError ErrorType = "prompt"
	// DownloadError is an error trying to download the update
	DownloadError ErrorType = "download"
	// ApplyError is an error applying the update
	ApplyError ErrorType = "apply"
	// VerifyError is an error verifing the update (signature or digest)
	VerifyError ErrorType = "verify"
	// RestartError is an error with the restart
	RestartError ErrorType = "restart"
)

func (t ErrorType) String() string {
	return string(t)
}

// Error is an update error with a type/category for reporting
type Error struct {
	errorType ErrorType
	source    error
}

// NewError constructs an Error from a source error
func NewError(errorType ErrorType, err error) Error {
	return Error{errorType: errorType, source: err}
}

// TypeString returns a unique short string to denote the error type
func (e Error) TypeString() string {
	return e.errorType.String()
}

// Error returns description for an UpdateError
func (e Error) Error() string {
	if e.source == nil {
		return fmt.Sprintf("Update Error (%s)", e.TypeString())
	}
	return fmt.Sprintf("Update Error (%s): %s", e.TypeString(), e.source.Error())
}

func cancelErrPtr(err error) *Error {
	uerr := NewError(CancelError, err)
	return &uerr
}

func promptErrPtr(err error) *Error {
	uerr := NewError(PromptError, err)
	return &uerr
}

func findErrPtr(err error) *Error {
	uerr := NewError(FindError, err)
	return &uerr
}

func downloadErrPtr(err error) *Error {
	uerr := NewError(DownloadError, err)
	return &uerr
}

func verifyErrPtr(err error) *Error {
	uerr := NewError(VerifyError, err)
	return &uerr
}

func applyErrPtr(err error) *Error {
	uerr := NewError(ApplyError, err)
	return &uerr
}

func restartErrPtr(err error) *Error {
	uerr := NewError(RestartError, err)
	return &uerr
}

func configErr(err error) Error {
	return NewError(ConfigError, err)
}
