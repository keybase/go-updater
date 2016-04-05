// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"github.com/keybase/client/go/logger"
	"github.com/keybase/go-updater"
)

type service struct {
	updater       *updater.Updater
	updateChecker *updater.UpdateChecker
	context       updater.Context
	log           logger.Logger
	ch            chan int
}

func newService(upd *updater.Updater, ctx updater.Context, log logger.Logger) *service {
	svc := service{}
	svc.updater = upd
	svc.context = ctx
	svc.log = log
	return &svc
}

func (s service) Start() {
	if s.updateChecker == nil {
		updateChecker := updater.NewUpdateChecker(s.updater, s.context, s.log)
		s.updateChecker = &updateChecker
	}
	s.updateChecker.Start()
}

func (s service) Run() int {
	s.Start()
	<-s.ch
	return 0
}
