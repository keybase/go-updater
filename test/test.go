// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// This is a test executable built and installed prior to test run, which is
// useful for testing some command.go functions.
func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Printf("Got SIGTERM, not exiting on purpose")
		// Don't exit on SIGTERM, so we can test timeout with SIGKILL
	}()
	fmt.Printf("Waiting for 10 seconds...")
	time.Sleep(10 * time.Second)
}
