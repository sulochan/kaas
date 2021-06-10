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
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

// CreateResult is the result of a Create operation
type CreateResult struct {
	gophercloud.Result
}

// Extract interprets a CreateResult as a Node.
func (r CreateResult) Extract() (*Node, error) {
	var s struct {
		Response struct {
			Node *Node `json:"node"`
		} `json:"response"`
	}
	err := r.ExtractInto(&s)
	return s.Response.Node, err
}

// method to determine if the call succeeded or failed.
type DeleteResult struct {
	gophercloud.ErrResult
}

// GetResult is the response from a Get operation. Call its Extract method to
// interpret it as a Node.
type GetResult struct {
	gophercloud.Result
}

// DeleteResult is the result from a Delete operation. Call its ExtractErr

// Extract interprets a GetResult as a Node.
func (r GetResult) Extract() (*Node, error) {
	var s struct {
		Node Node `json:"node"`
	}
	err := r.ExtractInto(&s)
	return &s.Node, err
}

// NodePage contains a single page of all Nodes return from a List
// operation. Use ExtractNodes to convert it into a slice of usable structs.
type NodePage struct {
	pagination.LinkedPageBase
}

// IsEmpty returns true if response contains no Node results.
func (r NodePage) IsEmpty() (bool, error) {
	nodes, err := ExtractNodes(r)
	return len(nodes) == 0, err
}

// NextPageURL uses the response's embedded link reference to navigate to the
// next page of results.
func (page NodePage) NextPageURL() (string, error) {
	var s struct {
		Links []gophercloud.Link `json:"links"`
	}
	err := page.ExtractInto(&s)
	if err != nil {
		return "", err
	}
	return gophercloud.ExtractNextURL(s.Links)
}

// ExtractNodes converts a page of List results into a slice of usable Node
// structs.
func ExtractNodes(r pagination.Page) ([]Node, error) {
	var s struct {
		Nodes []Node `json:"nodes"`
	}
	err := (r.(NodePage)).ExtractInto(&s)
	return s.Nodes, err
}

// Node represents a node behind a load balancer returned by the Cloud Load Balancer API.
type Node struct {
	// ID is the unique ID of a node
	ID uint64 `json:"id"`

	// The address of the node
	Address string `json:"address"`

	// The port for the node
	Port int32 `json:"port"`

	// Indicates if the node is ENABLED, DISABLED or DRAINING
	Condition string `json:"condition"`

	// The status of the node
	Status string `json:"status"`

	// Indicates the weight for the node
	Weight uint `json:"weight"`

	// Node typeof virtualIps for a load balancer
	Type string `json:"type"`
}
