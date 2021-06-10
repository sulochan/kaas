/*
Copyright 2021 Rackspace, Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use
this file except in compliance with the License.  You may obtain a copy of the
License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied.  See the License for the
specific language governing permissions and limitations under the License.
*/

package loadbalancers

import (
	"log"
	"strconv"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"

	"github.com/os-pc/gocloudlb/nodes"
	"github.com/os-pc/gocloudlb/virtualips"
)

// ListOptsBuilder allows extensions to add additional parameters to the
// List request.
type ListOptsBuilder interface {
	ToLoadBalancerListQuery() (string, error)
}

// ListOpts contain options filtering LoadBalancers returned from a call to List.
type ListOpts struct {
	// Status of the LoadBalancer
	Status string `q:"status"`
	// Address of a node attached to the LoadBalancer
	Node string `q:"nodeaddress"`
}

// ToLoadBalancerListQuery formats a ListOpts into a query string.
func (opts ListOpts) ToLoadBalancerListQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	return q.String(), err
}

func List(client *gophercloud.ServiceClient, opts ListOptsBuilder) pagination.Pager {
	url := client.ServiceURL("loadbalancers")
	if opts != nil {
		query, err := opts.ToLoadBalancerListQuery()
		if err != nil {
			return pagination.Pager{Err: err}
		}
		url += query
	}

	log.Printf("GET %s", url)

	return pagination.NewPager(client, url, func(r pagination.PageResult) pagination.Page {
		return LoadBalancerPage{pagination.LinkedPageBase{PageResult: r}}
	})
}

// Get returns data about a specific load balancer by its ID.
func Get(client *gophercloud.ServiceClient, id uint64) (r GetResult) {
	url := client.ServiceURL("loadbalancers", strconv.FormatUint(id, 10))
	log.Printf("GET %s", url)
	_, r.Err = client.Get(url, &r.Body, nil)
	return
}

// Delete deletes the specified loadbalancer ID.
func Delete(client *gophercloud.ServiceClient, id uint64) (r DeleteResult) {
	url := client.ServiceURL("loadbalancers", strconv.FormatUint(id, 10))
	log.Printf("DELETE %s", url)

	_, r.Err = client.Delete(url, &gophercloud.RequestOpts{})
	return
}

// CreateOpts contain the values necessary to create a loadbalancer
type CreateOpts struct {
	// Name is the name of the LoadBalancer.
	Name string `json:"name"`

	// Protocol of the service that is being load balanced
	Protocol string `json:"protocol"`

	// Port number for the service you are load balancing
	Port int32 `json:"port"`

	// Algorithm that defines how traffic should be directed between back-end nodes
	Algorithm string `json:"algorithm,omitempty"`

	// The list of virtualIps for a load balancer
	VirtualIps []virtualips.CreateOpts `json:"virtualIps"`

	Nodes []nodes.CreateOpts `json:"nodes"`
}

// Create creates a requested loadbalancer
func Create(client *gophercloud.ServiceClient, opts CreateOpts) (r CreateResult) {
	url := client.ServiceURL("loadbalancers")

	log.Printf("POST %s", url)

	var body = struct {
		LoadBalancer CreateOpts `json:"loadBalancer"`
	}{
		opts,
	}

	_, r.Err = client.Post(url, body, &r.Body, nil)
	return
}
