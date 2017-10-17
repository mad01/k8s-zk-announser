package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
)

const (
	eventCreate       = "create"
	eventUpdate       = "update"
	eventDelete       = "delete"
	serviceAnnotation = "zookeeper/path"
)

func newUpdaterEvent(eventType string, service *v1.Service) (*UpdaterEvent, error) {
	member := newZKMember()
	member.name = service.GetName()

	annotations := service.GetAnnotations()
	if path, ok := annotations[serviceAnnotation]; ok {
		member.path = path
	} else {
		return nil, fmt.Errorf("failed to find service annotation: %v on service: %v", serviceAnnotation, member.name)
	}

	event := UpdaterEvent{
		eventType: eventType,
		member:    member,
	}
	return &event, nil
}

// UpdaterEvent create/update/delete of zkmember
type UpdaterEvent struct {
	eventType string // create/update/delete
	member    *zkMember
}

func newUpdater() *Updater {
	updater := Updater{
		events: make(chan UpdaterEvent),
	}
	return &updater
}

// Updater event worker
type Updater struct {
	events chan UpdaterEvent
}

// Run starts to wait for events and executes them
func (u *Updater) Run(stopCh chan struct{}) {
	for {
		select {
		case event := <-u.events:
			log.Infof("process event: %v service: %v", event.eventType, event.member.name)
		case _ = <-stopCh:
			log.Info("stopping updater runner")
			return
		}
	}
}
