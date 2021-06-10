package models

import (
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/os-pc/gocloudlb/loadbalancers"
)

type Cluster struct {
	UUID         string `json:"uuid"`
	Name         string `json:"name"`
	Config       string `json:"config"`
	URL          string `json:"url"`
	LBNode       *loadbalancers.LoadBalancer
	Master       int     `json:"master"`
	MasterNodes  []*Node `json:"masternodes"`
	Worker       int     `json:"worker"`
	WorkerNodes  []*Node `json:"workernodes"`
	Etcd         int     `json:"etcd"`
	EtcdNodes    []*Node `json:"etcdnodes"`
	ExternalEtcd bool    `json:"externaletcd"`
	Nodes        []*Node `json:"nodes"`
	OSClient     *gophercloud.ServiceClient
	CreatedAt    time.Time `json:"createdat"`
	Deleted      int       `json:"deleted"`
	Status       string    `json:"status"`
	// accounted related info
	ProjectId string `json:"projectid"`
	CreatedBy string `json:"createdby"`
	Region    string `json:"region"`
}

type Public struct {
	Names []string
	Nodes []Node
}

type Node struct {
	Cluster    string
	IP         string
	InternalIP string
	Roles      []string
	Name       string
	Password   string
	UUID       string
	Type       string
}
