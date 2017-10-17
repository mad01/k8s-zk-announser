package main

import (
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
	conn *zk.Conn
}

// Conn (connect to zookeeper)
func (z *Zoo) Conn(server string) error {
	log.Infof("connecting to zk %v", server)
	c, _, err := zk.Connect([]string{server}, 10*time.Second) //*10)
	if err != nil {
		return err
	}
	z.conn = c
	log.Info("connected to zk %v", server)
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
func (z *Zoo) AddServiceMember(member *zkMember) (string, error) {
	err := z.createFullPath(member.path)
	if err != nil {
		return "", err
	}

	memberData, err := member.marshalJSON()
	if err != nil {
		return "", err
	}

	// TODO: change to set 0 to something else to support multiple
	path := fmt.Sprintf("%s/%s%s", member.path, memberPrefix, "0")

	log.Debugf("trying to add service member with path: %s", member.path)
	respPath, err := z.conn.Create(
		path,
		memberData,
		1,
		zk.WorldACL(zk.PermAll),
	)

	if err == zk.ErrNodeExists {
		return respPath, nil
	} else if err != nil {
		log.Errorf("failed to create service member in path: %s  err: %s ", member.path, err.Error())
		return "", err
	}
	log.Infof("added service member: %s with path: %s", member.name, member.path)
	return respPath, nil
}
