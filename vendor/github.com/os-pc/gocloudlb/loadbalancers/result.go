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
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"

	"github.com/os-pc/gocloudlb/nodes"
	"github.com/os-pc/gocloudlb/virtualips"
)

// CreateResult is the result of a Create operation
type CreateResult struct {
	gophercloud.Result
}

// Extract interprets a CreateResult as a LoadBalancer.
func (r CreateResult) Extract() (*LoadBalancer, error) {
	var s struct {
		LoadBalancer *LoadBalancer `json:"loadBalancer"`
	}
	err := r.ExtractInto(&s)
	return s.LoadBalancer, err
}

// method to determine if the call succeeded or failed.
type DeleteResult struct {
	gophercloud.ErrResult
}

// GetResult is the response from a Get operation. Call its Extract method to
// interpret it as a LoadBalancer.
type GetResult struct {
	gophercloud.Result
}

// DeleteResult is the result from a Delete operation. Call its ExtractErr

// Extract interprets a GetResult as a LoadBalancer.
func (r GetResult) Extract() (*LoadBalancer, error) {
	var s struct {
		LoadBalancer LoadBalancer `json:"loadBalancer"`
	}
	err := r.ExtractInto(&s)
	return &s.LoadBalancer, err
}

// LoadBalancerPage contains a single page of all LoadBalancers return from a List
// operation. Use ExtractLoadBalancers to convert it into a slice of usable structs.
type LoadBalancerPage struct {
	pagination.LinkedPageBase
}

// IsEmpty returns true if response contains no LoadBalancer results.
func (r LoadBalancerPage) IsEmpty() (bool, error) {
	loadbalancers, err := ExtractLoadBalancers(r)
	return len(loadbalancers) == 0, err
}

// NextPageURL uses the response's embedded link reference to navigate to the
// next page of results.
func (page LoadBalancerPage) NextPageURL() (string, error) {
	var s struct {
		Links []gophercloud.Link `json:"links"`
	}
	err := page.ExtractInto(&s)
	if err != nil {
		return "", err
	}
	return gophercloud.ExtractNextURL(s.Links)
}

// ExtractLoadBalancers converts a page of List results into a slice of usable LoadBalancer
// structs.
func ExtractLoadBalancers(r pagination.Page) ([]LoadBalancer, error) {
	var s struct {
		LoadBalancers []LoadBalancer `json:"loadBalancers"`
	}
	err := (r.(LoadBalancerPage)).ExtractInto(&s)
	return s.LoadBalancers, err
}

// LoadBalancer represents a load balancer returned by the Cloud Load Balancer API.
type LoadBalancer struct {
	// ID is the unique ID of a load balancer.
	ID uint64 `json:"id"`

	// Name of the loadbalancer
	Name string `json:"name"`

	// Protocol of the service that is being load balanced
	Protocol string `json:"protocol"`

	// Port number for the service you are load balancing
	Port int32 `json:"port"`

	// Algorithm that defines how traffic should be directed between back-end nodes
	Algorithm string `json:"algorithm"`

	// The status of the load balancer
	Status string `json:"status"`

	// The timeout value for the load balancer and communications with its nodes
	Timeout uint `json:"timeout"`

	// The list of virtualIps for a load balancer
	VirtualIps []virtualips.VirtualIp `json:"virtualIps"`

	// Nodes servicing the requests
	Nodes []nodes.Node `json:"nodes"`

	// Created is the date when the load balancer was created.
	Created struct {
		Time string `json:"time"`
	} `json:"created"`

	// Updated is the date when the load balancer was updated.
	Updated struct {
		Time string `json:"time"`
	} `json:"updated"`
}
