// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package util

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
)

// JoinPredicate joins strings with predicate
func JoinPredicate(arr []string, delimeter string, f func(s string) bool) string {
	arrNew := make([]string, 0, len(arr))
	for _, s := range arr {
		if f(s) {
			arrNew = append(arrNew, s)
		}
	}
	return strings.Join(arrNew, delimeter)
}

// RandString returns random (base32) string with prefix.
func RandString(prefix string, numbytes int) (string, error) {
	buf, err := RandBytes(numbytes)
	if err != nil {
		return "", err
	}
	str := prefix + base32.StdEncoding.EncodeToString(buf)
	return str, nil
}

var randRead = rand.Read

// RandBytes returns random bytes of length
func RandBytes(length int) ([]byte, error) {
	buf := make([]byte, length)
	if _, err := randRead(buf); err != nil {
		return nil, err
	}
	return buf, nil
}
