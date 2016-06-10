// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
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
		var iface interface{}
		err := json.Unmarshal([]byte(os.Args[3]), &iface)
		if err != nil {
			fmt.Printf("Error unmarshaling: %v", err)
			return
		}
		promptArgs := iface.(map[string]interface{})
		echoToFile(flag.Arg(1), promptArgs["outPath"].(string))
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

func echoToFile(s string, path string) {
	err := ioutil.WriteFile(path, []byte(s), 0777)
	if err != nil {
		log.Fatal("Error writing file")
		return
	}
}
