package handler

import (
	"net/http"
	"sync"

	"github.com/winston-ci/redgreen/api/builds"
	"github.com/winston-ci/redgreen/logbuffer"
)

type Handler struct {
	peerAddr string
	proleURL string

	builds      map[string]*builds.Build
	buildsMutex *sync.RWMutex

	logs      map[string]*logbuffer.LogBuffer
	logsMutex *sync.RWMutex

	bits      map[string]BitsSession
	bitsMutex *sync.RWMutex
}

type BitsSession struct {
	bits        chan *http.Request
	servingBits *sync.WaitGroup
}

func NewHandler(peerAddr string, proleURL string) *Handler {
	return &Handler{
		peerAddr: peerAddr,
		proleURL: proleURL,

		builds:      make(map[string]*builds.Build),
		buildsMutex: new(sync.RWMutex),

		logs:      make(map[string]*logbuffer.LogBuffer),
		logsMutex: new(sync.RWMutex),

		bits:      make(map[string]BitsSession),
		bitsMutex: new(sync.RWMutex),
	}
}
