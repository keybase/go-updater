// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater"
)

type service struct {
	updater       *updater.Updater
	updateChecker *updater.UpdateChecker
	context       updater.Context
	log           logging.Logger
	ch            chan int
}

func newService(upd *updater.Updater, context updater.Context, log logging.Logger) *service {
	svc := service{
		updater: upd,
		context: context,
		log:     log,
		ch:      make(chan int),
	}
	return &svc
}

func (s *service) Start() {
	if s.updateChecker == nil {
		updateChecker := updater.NewUpdateChecker(s.updater, s.context, s.log)
		s.updateChecker = &updateChecker
	}
	s.updateChecker.Start()
	// If you want to check on service startup
	//s.updateChecker.Check()
}

func (s *service) Run() {
	s.Start()
	<-s.ch
}

func (s *service) Quit() {
	s.ch <- 0
}
