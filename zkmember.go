package main

import (
	"encoding/json"
)

//{
//  "status": "ALIVE",
//  "additionalEndpoints": {
//    "health": {
//      "host": "10.19.64.35",
//      "port": 10764
//    },
//    "http": {
//      "host": "10.19.64.35",
//      "port": 10691
//    }
//  },
//  "serviceEndpoint": {
//    "host": "10.19.64.35",
//    "port": 10691
//  },
//  "shard": 0
//}

// possible endpoint statuses. Currently only concerned with ALIVE.
const (
	statusDead     = "DEAD"
	statusStarting = "STARTING"
	statusAlive    = "ALIVE"
	statusStopping = "STOPPING"
	statusStopped  = "STOPPED"
	statusWarning  = "WARNING"
	statusUnknown  = "UNKNOWN"
)

type zkMemberUnite struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// Endpoints for zookeeper members
type Endpoints map[string]zkMemberUnite

type zkMember struct {
	name   string
	path   string // zookeeper path
	prefix string

	Status              string        `json:"status"` // set to ALIVE
	AdditionalEndpoints Endpoints     `json:"additionalEndpoints"`
	ServiceEndpoint     zkMemberUnite `json:"serviceEndpoint"`
	Shard               int           `json:"shard"`
}

func (z *zkMember) addAdditionalEndpoints(name, addr string, port int) {
	z.AdditionalEndpoints[name] = zkMemberUnite{Host: addr, Port: port}
}

func (z *zkMember) addServiceEndpoint(name, addr string, port int) {
	z.ServiceEndpoint = zkMemberUnite{Host: addr, Port: port}
	z.AdditionalEndpoints[name] = zkMemberUnite{Host: addr, Port: port}
}

func (z *zkMember) marshalJSON() ([]byte, error) {
	data, err := json.Marshal(z)
	if err != nil {
		var b []byte
		return b, err
	}
	return data, err
}

func (z *zkMember) unmarshalJSON(bytebuff []byte) (*zkMember, error) {
	member := newZKMember()
	if err := json.Unmarshal(bytebuff, &member); err != nil {
		return member, err
	}
	return member, nil
}

func (z *zkMember) anyEndpoints() bool {
	if len(z.AdditionalEndpoints) >= 1 && (z.ServiceEndpoint.Host != "" && z.ServiceEndpoint.Port != 0) {
		return true
	}
	return false
}

// newZLKMember returns instance of new member
func newZKMember() *zkMember {
	member := zkMember{
		AdditionalEndpoints: make(Endpoints),
		Status:              statusAlive,
	}
	return &member
}
