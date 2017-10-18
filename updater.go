package main

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
)

const (
	eventCreate               = "create"
	eventUpdate               = "update"
	eventDelete               = "delete"
	serviceAnnotationPath     = "service.announser/zookeeper-path"
	serviceAnnotationPortName = "service.announser/portname"
)

func checkRequiredServiceFieldsExists(service *v1.Service) error {
	annotations := service.GetAnnotations()
	if _, ok := annotations[serviceAnnotationPortName]; !ok {
		return fmt.Errorf("missing annotation %v", serviceAnnotationPortName)
	}
	if _, ok := annotations[serviceAnnotationPath]; !ok {
		return fmt.Errorf("missing annotation %v", serviceAnnotationPath)
	}
	if service.Spec.Type != "LoadBalancer" {
		return fmt.Errorf("only type LoadBalancer supported. %v not supported yet", service.Spec.Type)
	}
	return nil
}

func getServicePortByName(name string, service *v1.Service) *v1.ServicePort {
	for _, port := range service.Spec.Ports {
		if port.Name == name {
			return &port
		}
	}
	return nil
}

func getServiceAddr(service *v1.Service) string {
	for _, val := range service.Status.LoadBalancer.Ingress {
		if val.Hostname != "" {
			return val.Hostname
		} else if val.IP != "" {
			return val.IP
		}
	}
	return ""
}

func newUpdaterEvent(eventType string, service *v1.Service) (*UpdaterEvent, error) {
	err := checkRequiredServiceFieldsExists(service)
	if err != nil {
		return nil, fmt.Errorf("error service %v, err: %v", service.GetName(), err.Error())
	}

	annotations := service.GetAnnotations()
	member := newZKMember()
	member.path = annotations[serviceAnnotationPath]
	member.name = service.GetName()
	member.prefix = service.GetResourceVersion()

	portname := annotations[serviceAnnotationPortName]
	port := getServicePortByName(portname, service)
	if port == nil {
		return nil, fmt.Errorf("service named missing port")
	}
	serviceAddr := getServiceAddr(service)
	if serviceAddr == "" {
		return nil, fmt.Errorf("missing LoadBalancer hostname or ip will retry")
	}
	member.addServiceEndpoint(
		portname,
		serviceAddr,
		int(port.Port),
	)

	event := UpdaterEvent{
		eventType:  eventType,
		member:     member,
		retryCount: 5,
		retryWait:  5 * time.Second,
	}
	return &event, nil
}

// UpdaterEvent create/update/delete of zkmember
type UpdaterEvent struct {
	eventType  string // create/update/delete
	member     *zkMember
	retryCount int
	retryWait  time.Duration
}

func newUpdater(zookeeperAddr string) *Updater {
	updater := Updater{
		events:        make(chan UpdaterEvent),
		zookeeperAddr: zookeeperAddr,
	}
	updater.zookeeper.Init()
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
			log.Debugf("process event: %v service: %v", event.eventType, event.member.name)
			switch event.eventType {
			case eventCreate:
				log.Debugf("create event")
				err := u.zookeeper.AddServiceMember(event.member)
				if err != nil {
					log.Errorf("failed to create member: %v %v", event.member.name, err.Error())
				}
			case eventUpdate:
				log.Debugf("update event not implemented")
				err := u.zookeeper.AddServiceMember(event.member)
				if err != nil {
					log.Errorf("failed to update member: %v %v", event.member.name, err.Error())
				}
			case eventDelete:
				log.Debugf("delete event")
				err := u.zookeeper.DeleteServiceMember(event.member)
				if err != nil {
					log.Errorf("failed to delete member: %v %v", event.member.name, err.Error())
				}
			}
		case _ = <-stopCh:
			log.Info("stopping updater runner")
			return
		}
	}
}
