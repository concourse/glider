package builds

import "time"

type Build struct {
	Guid        string            `json:"guid"`
	CreatedAt   time.Time         `json:"created_at"`
	Image       string            `json:"image"`
	Script      string            `json:"script"`
	Environment map[string]string `json:"environment"`
	Status      string            `json:"status"`
}
