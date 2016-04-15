// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package updater

import (
	"fmt"
	"testing"
)

func TestNewUpdateError(t *testing.T) {
	err := NewUpdateError(UpdatePromptError, fmt.Errorf("There was an error prompting"))
	if err.Error() != "Update Error (prompt): There was an error prompting" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}
