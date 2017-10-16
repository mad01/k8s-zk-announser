package main

import (
	"encoding/json"
)

// {
//  "status": "ALIVE",
//  "additionalEndpoints": {
//    "admin": {
//      "host": "10.19.64.35",
//      "port": 10764
//    },
//    "websocket": {
//      "host": "10.19.64.35",
//      "port": 10691
//    },
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

type zkMemberUnite struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// Endpoints for zk members
type Endpoints map[string]zkMemberUnite

type zkMember struct {
	serviceName string
	role        string // game/title
	env         string // prod/int/test/stage
	path        string // zk path

	zkBasePath          string
	Status              string    `json:"status"` // set to ALIVE
	AdditionalEndpoints Endpoints `json:"additionalEndpoints"`
	ServiceEndpoint     Endpoints `json:"serviceEndpoint"`
	Shard               int       `json:"shard"`
}

func (z *zkMember) Init() {
	z.AdditionalEndpoints = make(Endpoints)
	z.ServiceEndpoint = make(Endpoints)
}

func (z *zkMember) addAdditionalEndpoints(name string, unit zkMemberUnite) {
	z.AdditionalEndpoints[name] = unit
}

func (z *zkMember) addServiceEndpoint(name string, unit zkMemberUnite) {
	z.ServiceEndpoint[name] = unit
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
	if len(z.AdditionalEndpoints)|len(z.ServiceEndpoint) >= 1 {
		return true
	}
	return false
}

// newZLKMember returns instance of new member
func newZKMember() *zkMember {
	var member zkMember
	member.Init()
	return &member
}
