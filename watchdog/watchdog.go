// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package watchdog

import (
	"os"
	"os/exec"
	"time"

	"github.com/keybase/go-updater/process"
)

// ExitOn describes when a program should exit (not-restart)
type ExitOn string

const (
	// ExitOnNone means the program should always be restarted
	ExitOnNone ExitOn = ""
	// ExitOnSuccess means the program should only restart if errored
	ExitOnSuccess ExitOn = "success"
)

// Program is a program at path with arguments
type ProgramNormal struct {
	Path   string
	Args   []string
	ExitOn ExitOn
}

func (p *ProgramNormal) GetPath() string {
	return p.Path
}

func (p *ProgramNormal) GetArgs() []string {
	return p.Args
}

func (p *ProgramNormal) GetExitOn() ExitOn {
	return p.ExitOn
}

func (p *ProgramNormal) DoStop(ospid int, log Log) {
	StopMatching(p.Path, ospid, log)
}

// Program is a program at path with arguments
type Program interface {
	GetPath() string
	GetArgs() []string
	GetExitOn() ExitOn
	DoStop(ospid int, log Log)
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
	// Terminate any existing programs that we are supposed to monitor
	terminateExisting(programs, log)

	// Start monitoring all the programs
	watchPrograms(programs, restartDelay, log)
	return nil
}

func StopMatching(path string, ospid int, log Log) {
	matcher := process.NewMatcher(path, process.PathEqual, log)
	matcher.ExceptPID(ospid)
	log.Infof("Terminating %s", path)
	process.TerminateAll(matcher, time.Second, log)
}

func terminateExisting(programs []Program, log Log) {
	// Terminate any monitored processes
	ospid := os.Getpid()
	log.Infof("Terminating any existing programs we will be monitoring")
	for _, program := range programs {
		program.DoStop(ospid, log)
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
		log.Infof("Starting %#v", program)
		cmd := exec.Command(program.GetPath(), program.GetArgs()...)
		err := cmd.Run()
		if err != nil {
			log.Errorf("Error running program: %q; %s", program, err)
		} else {
			log.Infof("Program finished: %q", program)
			if program.GetExitOn() == ExitOnSuccess {
				log.Infof("Program configured to exit on success, not restarting")
				break
			}
		}
		log.Infof("Program ran for %s", time.Since(start))
		if time.Since(start) < restartDelay {
			log.Infof("Waiting %s before trying to start command again", restartDelay)
			time.Sleep(restartDelay)
		}
	}
}
