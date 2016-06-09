// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// This is a test executable built and installed prior to test run, which is
// useful for testing some command.go functions.
func main() {
	flag.Parse()
	var arg = flag.Arg(0)

	switch arg {
	case "noexit":
		noexit()
	case "output":
		output()
	case "echo":
		echo(flag.Arg(1))
	case "echoToFile":
		// Trying to parse 2 separate json argument objects is too gross -
		// just pick out the pathname
		for _, a := range os.Args[2:] {
			if idx := strings.Index(a, "outPathName"); idx != -1 {
				tmp := a[idx+11:]
				tmp = strings.TrimLeft(tmp, ":\", ")
				echoToFile(flag.Arg(1), tmp[:strings.IndexAny(tmp, ",}")-1])
				return
			}
		}
		echoToFile(flag.Arg(1), flag.Arg(2))
	case "version":
		echo("1.2.3-400+cafebeef")
	case "err":
		log.Fatal("Error")
	case "sleep":
		time.Sleep(10 * time.Second)
	default:
		log.Printf("test")
	}
}

func noexit() {
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

func output() {
	fmt.Fprintln(os.Stdout, "stdout output")
	fmt.Fprintln(os.Stderr, "stderr output")
}

func echo(s string) {
	fmt.Fprintln(os.Stdout, s)
}

func echoToFile(s string, pathName string) {
	f, err := os.OpenFile(pathName, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		fmt.Printf("\n --- ERROR opening file: %v ---\n", err)
		return
	}
	defer f.Close()
	f.WriteString(s)
}
