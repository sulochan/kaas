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

package nodes

import (
	"log"
	"strconv"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
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
	// Name of the LoadBalancer
	Name string `q:"name"`
}

// ToLoadBalancerListQuery formats a ListOpts into a query string.
func (opts ListOpts) ToLoadBalancerListQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	return q.String(), err
}

func List(client *gophercloud.ServiceClient, lbID uint64, opts ListOptsBuilder) pagination.Pager {
	url := client.ServiceURL("loadbalancers", strconv.FormatUint(lbID, 10), "nodes")
	if opts != nil {
		query, err := opts.ToLoadBalancerListQuery()
		if err != nil {
			return pagination.Pager{Err: err}
		}
		url += query
	}

	log.Printf("GET %s", url)

	return pagination.NewPager(client, url, func(r pagination.PageResult) pagination.Page {
		return NodePage{pagination.LinkedPageBase{PageResult: r}}
	})
}

// Get returns data about a specific node by its ID.
func Get(client *gophercloud.ServiceClient, lbID uint64, id uint64) (r GetResult) {
	url := client.ServiceURL("loadbalancers", strconv.FormatUint(lbID, 10), "nodes", strconv.FormatUint(id, 10))
	log.Printf("GET %s", url)
	_, r.Err = client.Get(url, &r.Body, nil)
	return
}

// Delete deletes the specified node ID.
func Delete(client *gophercloud.ServiceClient, lbID uint64, id uint64) (r DeleteResult) {
	url := client.ServiceURL("loadbalancers", strconv.FormatUint(lbID, 10), "nodes", strconv.FormatUint(id, 10))
	log.Printf("DELETE %s", url)

	_, r.Err = client.Delete(url, &gophercloud.RequestOpts{
		JSONResponse: &r.Body,
	})
	return
}

// CreateOpts contain the values necessary to create a node
type CreateOpts struct {
	// The address of the node
	Address string `json:"address"`

	// The port for the node
	Port int32 `json:"port"`

	// Indicates if the node is ENABLED, DISABLED, or DRAINING
	Condition string `json:"condition"`

	// Indicates the weight for the node
	Weight uint `json:"weight,omitempty"`

	// Node type
	Type string `json:"type,omitempty"`
}

// Create creates a requested node
func Create(client *gophercloud.ServiceClient, lbID uint64, opts []CreateOpts) (r CreateResult) {
	url := client.ServiceURL("loadbalancers", strconv.FormatUint(lbID, 10), "nodes")

	log.Printf("POST %s", url)

	var body = struct {
		Node []CreateOpts `json:"nodes"`
	}{
		opts,
	}

	_, r.Err = client.Post(url, body, &r.Body, nil)
	return
}
