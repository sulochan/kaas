package api

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Node is the structure of the computer to be registered
type Node struct {
	UUID           string `json:"uuid"`
	Hostname       string `json:"hostname"`
	IPAddress      string `json:"ip_address"`
	OS             string `json:"os_name"`
	OSVersion      string `json:"os_version"`
	OSArchitecture string `json:"os_architecture"`
	ProjectId      string `json:"project_id"`
}

// RegisterNode is the handler for POST /api/register
func RegisterNode(w http.ResponseWriter, r *http.Request) {
	log.Info("RegisterNode() called")
}

// GetNextJob is the handler for GET /api/get_next_job
func GetNextJob(w http.ResponseWriter, r *http.Request) {
	log.Info("GetNextJob() called")
}

// UpdateJob is the handler for GET /api/update_job
func UpdateJob(w http.ResponseWriter, r *http.Request) {
	log.Info("UpdateJob() called")
}
