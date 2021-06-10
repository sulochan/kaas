package gocloudlb

import (
	"github.com/gophercloud/gophercloud"
)

func NewLB(client *gophercloud.ProviderClient, eo gophercloud.EndpointOpts) (*gophercloud.ServiceClient, error) {

	serviceType := "rax:load-balancer"
	eo.ApplyDefaults(serviceType)
	url, err := client.EndpointLocator(eo)
	if err != nil {
		return nil, err
	}
	return &gophercloud.ServiceClient{ProviderClient: client, Endpoint: url, Type: serviceType}, nil
}
