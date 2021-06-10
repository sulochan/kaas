package models

type AuthOpts struct {
	Type string `json:"type"`
	Username string `json:"username"`
	ProjectId string `json:"projectId"`
	Token     string `json:"token"`
	Password  string `json:"password"`
}