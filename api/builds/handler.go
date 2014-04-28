package builds

import "sync"

type Handler struct {
	peerAddr string
	proleURL string

	builds      map[string]*Build
	buildsMutex *sync.RWMutex
}

func NewHandler(peerAddr string, proleURL string) *Handler {
	return &Handler{
		peerAddr: peerAddr,
		proleURL: proleURL,

		builds:      make(map[string]*Build),
		buildsMutex: new(sync.RWMutex),
	}
}
