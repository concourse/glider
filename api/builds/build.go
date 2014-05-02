package builds

import (
	"net/http"
	"sync"
	"time"

	"github.com/winston-ci/redgreen/logbuffer"
)

type Build struct {
	Guid      string      `json:"guid"`
	CreatedAt time.Time   `json:"created_at"`
	Image     string      `json:"image"`
	Path      string      `json:"path"`
	Script    string      `json:"script"`
	Env       [][2]string `json:"env"`
	Status    string      `json:"status"`

	bits        chan *http.Request
	servingBits *sync.WaitGroup

	logBuffer *logbuffer.LogBuffer
}

type BuildResult struct {
	Status string `json:"status"`
}
