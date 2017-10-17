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

func newUpdater(zookeeperAddr string) *Updater {
	updater := Updater{
		events:        make(chan UpdaterEvent),
		zookeeperAddr: zookeeperAddr,
	}
	return &updater
}

// Updater event worker
type Updater struct {
	events        chan UpdaterEvent
	zookeeper     Zoo
	zookeeperAddr string
}

// Run starts to wait for events and executes them
func (u *Updater) Run(stopCh chan struct{}) {
	log.Info("Starting Updater")
	err := u.zookeeper.Conn(u.zookeeperAddr)
	if err != nil {
		log.Errorf("failed to connect to zookeeper: %v", err.Error())
		close(stopCh)
	}

	for {
		select {
		case event := <-u.events:
			log.Infof("process event: %v service: %v", event.eventType, event.member.name)
			switch event.eventType {
			case eventCreate:
				log.Debugf("create event")
				path, err := u.zookeeper.AddServiceMember(event.member)
				if err != nil {
					log.Errorf("failed to create member: %v err: %v", event.member.name, err.Error())
				}
				log.Infof("created member %v in path :%v", event.member.name, path)
			case eventUpdate:
				log.Debugf("update event")
				path, err := u.zookeeper.AddServiceMember(event.member)
				if err != nil {
					log.Errorf("failed to update member: %v err: %v", event.member.name, err.Error())
				}
				log.Infof("updated member %v in path :%v", event.member.name, path)
			case eventDelete:
				log.Debugf("delete event")
			}
		case _ = <-stopCh:
			log.Info("stopping updater runner")
			return
		}
	}
}
