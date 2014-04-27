package builds

import (
	"net/http"
	"sync"
	"time"
)

type Build struct {
	Guid        string            `json:"guid"`
	CreatedAt   time.Time         `json:"created_at"`
	Image       string            `json:"image"`
	Script      string            `json:"script"`
	Environment map[string]string `json:"environment"`
	Status      string            `json:"status"`

	bits        chan *http.Request `json:"-"`
	servingBits sync.WaitGroup
}
