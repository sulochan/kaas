package models

// Job is the structure of the job to be executed
type Job struct {
	UUID       string `json:"uuid"`
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Error      string `json:"error"`
	Output     string `json:"output"`
	Category   string `json:"category"`
	Command    string `json:"command"`
	Data       string `json:"data"`
}
