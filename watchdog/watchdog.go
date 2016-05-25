// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package watchdog

import (
	"os"
	"os/exec"
	"time"

	"github.com/kardianos/osext"
	"github.com/keybase/go-updater/process"
)

// Program is a program at path with arguments
type Program struct {
	Path string
	Args []string
}

// Log is the logging interface for the watchdog package
type Log interface {
	Debugf(s string, args ...interface{})
	Infof(s string, args ...interface{})
	Warningf(s string, args ...interface{})
	Errorf(s string, args ...interface{})
}

// Watch monitors programs and restarts them if they aren't running
func Watch(programs []Program, restartDelay time.Duration, log Log) error {
	execPath, err := osext.Executable()
	if err != nil {
		return err
	}

	// Terminate any existing watchdogs (except for ourself)
	if err := terminateWatchdogs(execPath, time.Second, log); err != nil {
		return err
	}
	// Terminate any existing programs that we are supposed to monitor
	terminateExisting(programs, log)

	// Start monitoring all the programs
	watchPrograms(programs, restartDelay, log)
	return nil
}

func terminateWatchdogs(execPath string, killDelay time.Duration, log Log) error {
	procs, err := process.FindProcesses(process.NewMatcher(execPath, process.PathEqual, log), 0, 0, log)
	if err != nil {
		return err
	}
	ospid := os.Getpid()
	log.Infof("This process is running as PID %d", ospid)
	for _, proc := range procs {
		pid := proc.Pid()
		if pid != ospid {
			log.Warningf("There is another watchdog running, terminating")
			if err := process.TerminatePID(pid, killDelay, log); err != nil {
				log.Warningf("Error trying to terminate another supervisor")
			}
		}
	}
	return nil
}

func terminateExisting(programs []Program, log Log) {
	// Terminate any monitored processes
	ospid := os.Getpid()
	log.Infof("Terminating any existing programs we will be monitoring")
	for _, program := range programs {
		matcher := process.NewMatcher(program.Path, process.PathEqual, log)
		matcher.ExceptPID(ospid)
		process.TerminateAll(matcher, time.Second, log)
	}
}

func watchPrograms(programs []Program, delay time.Duration, log Log) {
	for _, program := range programs {
		go watchProgram(program, delay, log)
	}
}

// watchProgram will monitor a program and restart it if it exits.
// This method will run forever.
func watchProgram(program Program, restartDelay time.Duration, log Log) {
	for {
		start := time.Now()
		log.Infof("Starting %q", program)
		cmd := exec.Command(program.Path, program.Args...)
		err := cmd.Run()
		if err != nil {
			log.Errorf("Error running program: %q; %s", program, err)
		} else {
			log.Infof("Program finished: %q", program)
		}
		if time.Since(start) < time.Minute {
			log.Infof("Waiting %s before trying to start command again", restartDelay)
			time.Sleep(restartDelay)
		}
	}
}
