package main

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type Keychains struct {
	Org *Keychain
	Com *Keychain
}

func (ks Keychains) Update() {
	var err error

	if _, err = ks.Org.Update(); err != nil {
		log.WithError(err).WithField("keychain", "org").Error("could not update keychain")
	}

	if _, err = ks.Com.Update(); err != nil {
		log.WithError(err).WithField("keychain", "com").Error("could not update keychain")
	}
}

func (ks Keychains) Watch(d time.Duration, handler func(*Keychain)) {
	go ks.Org.Watch(d, handler)
	go ks.Com.Watch(d, handler)
}
