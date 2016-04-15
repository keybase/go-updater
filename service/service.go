// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import "github.com/keybase/go-logging"

type service struct {
	log logging.Logger
	ch  chan int
}

func newService(log logging.Logger) *service {
	svc := service{}
	return &svc
}

func (s service) Start() {
	// TODO:
}

func (s service) Run() int {
	s.Start()
	<-s.ch
	return 0
}
