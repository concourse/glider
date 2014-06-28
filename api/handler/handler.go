package handler

import (
	"net/http"
	"sync"

	"github.com/concourse/glider/api/builds"
	"github.com/concourse/logbuffer"
	"github.com/pivotal-golang/lager"
)

type Handler struct {
	logger lager.Logger

	peerAddr   string
	turbineURL string

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

func NewHandler(logger lager.Logger, peerAddr string, turbineURL string) *Handler {
	return &Handler{
		logger: logger,

		peerAddr:   peerAddr,
		turbineURL: turbineURL,

		builds:      make(map[string]*builds.Build),
		buildsMutex: new(sync.RWMutex),

		logs:      make(map[string]*logbuffer.LogBuffer),
		logsMutex: new(sync.RWMutex),

		bits:      make(map[string]BitsSession),
		bitsMutex: new(sync.RWMutex),
	}
}
