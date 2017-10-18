package main

import (
	"fmt"
	"path"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	log "github.com/sirupsen/logrus"
)

var (
	memberPrefix = "member_"
)

// Zoo zookeeper main struct
type Zoo struct {
	conn   *zk.Conn
	active *activeMembers
}

// Init the active memebers map
func (z *Zoo) Init() {
	z.active = newActiveMembers()
}

// Conn (connect to zookeeper)
func (z *Zoo) Conn(server string) error {
	c, _, err := zk.Connect([]string{server}, 10*time.Second) //*10)
	if err != nil {
		return err
	}
	z.conn = c
	return nil
}

func (z *Zoo) splitPaths(fullPath string) []string {
	var parts []string

	var last string
	for fullPath != "/" {
		fullPath, last = path.Split(path.Clean(fullPath))
		parts = append(parts, last)
	}

	// parts are in reverse order, put back together
	// into set of subdirectory paths
	result := make([]string, 0, len(parts))
	base := ""
	for i := len(parts) - 1; i >= 0; i-- {
		base += "/" + parts[i]
		result = append(result, base)
	}

	return result
}

// createFullPath makes sure all the znodes are created for the parent directories
func (z *Zoo) createFullPath(path string) error {
	paths := z.splitPaths(path)
	for _, key := range paths {
		log.Debugf("create path key: %s", key)
		_, err := z.conn.Create(key, nil, 0, zk.WorldACL(zk.PermAll))
		if err != nil && err != zk.ErrNodeExists {
			log.Errorf("error creating full zk path: %s\n", err.Error())
			return err
		}
	}

	return nil
}

// AddServiceMember add new zk member
func (z *Zoo) AddServiceMember(member *zkMember) error {
	if !member.anyEndpoints() {
		return fmt.Errorf("failed to add no service endpoints")
	}
	if z.active.keyIn(member.name) {
		return fmt.Errorf("will not add member exists in zk")
	}
	err := z.createFullPath(member.path)
	if err != nil {
		return err
	}

	memberData, err := member.marshalJSON()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%s", member.path, memberPrefix)

	log.Debugf("trying to add service member with path: %s", member.path)
	respPath, err := z.conn.Create(
		path,
		memberData,
		zk.FlagEphemeral|zk.FlagSequence,
		zk.WorldACL(zk.PermAll),
	)

	if err == zk.ErrNodeExists {
		return nil
	} else if err != nil {
		log.Errorf("failed to create service member in path: %s  err: %s ", member.path, err.Error())
		return err
	}
	log.Infof("added service member: %s with path: %s", member.name, respPath)
	z.active.add(member.name, respPath)
	return nil
}

// DeleteServiceMember delete member
func (z *Zoo) DeleteServiceMember(member *zkMember) error {
	path := z.active.get(member.name)
	if path == "" {
		return fmt.Errorf("Missing path for service %v", member.name)
	}
	err := z.conn.Delete(path, 0)
	if err != nil {
		return fmt.Errorf("failed to delete service member in path %v err: %v", member.path, err.Error())
	}
	log.Infof("deleted member: %v", path)
	z.active.delete(member.name)
	return nil
}
