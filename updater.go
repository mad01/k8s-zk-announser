package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

const (
	updaterEventdCreate = "create"
	updaterEventUpdate  = "update"
	updaterEventDelete  = "delete"
)

func newUpdaterEvent(eventType string, member *zkMember) *UpdaterEvent {
	event := UpdaterEvent{
		eventType: eventType,
		member:    member,
	}
	return &event
}

// UpdaterEvent create/update/delete of zkmember
type UpdaterEvent struct {
	eventType string // create/update/delete
	member    *zkMember
}

func newUpdater() *Updater {
	updater := Updater{
		events: make(chan UpdaterEvent, 100),
	}
	return &updater
}

// Updater event worker
type Updater struct {
	events chan UpdaterEvent
}

// Run starts to wait for events and executes them
func (u *Updater) Run(events chan UpdaterEvent, stopCh chan struct{}) {
	for {
		select {
		case event := <-events:
			log.Infof("process event: %v", event.eventType)
			fmt.Println("act on event")
		case _ = <-stopCh:
			fmt.Println("stopping updater runner")
			return
		}
	}
}
