package builds

import "time"

type Build struct {
	Guid      string      `json:"guid"`
	CreatedAt time.Time   `json:"created_at"`
	Image     string      `json:"image"`
	Path      string      `json:"path"`
	Script    string      `json:"script"`
	Env       [][2]string `json:"env"`
	Status    string      `json:"status,omitempty"`
}

type BuildResult struct {
	Status string `json:"status"`
}
