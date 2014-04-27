package builds

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/nu7hatch/gouuid"
	"github.com/rcrowley/go-tigertonic"
)

type Handler struct {
	proleURL string

	builds      map[string]*Build
	buildsMutex *sync.RWMutex
}

func NewHandler(proleURL string) *Handler {
	return &Handler{
		proleURL: proleURL,

		builds:      make(map[string]*Build),
		buildsMutex: new(sync.RWMutex),
	}
}

func (handler *Handler) PostHandler() http.Handler {
	return tigertonic.Marshaled(handler.post)
}

func (handler *Handler) GetHandler() http.Handler {
	return tigertonic.Marshaled(handler.get)
}

func (handler *Handler) post(url *url.URL, header http.Header, build *Build) (int, http.Header, *Build, error) {
	err := handler.validateBuild(build)
	if err != nil {
		return http.StatusBadRequest, nil, nil, err
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	build.Guid = uuid.String()
	build.CreatedAt = time.Now()

	log.Println("registering", build.Guid)

	handler.buildsMutex.Lock()
	handler.builds[build.Guid] = build
	handler.buildsMutex.Unlock()

	return http.StatusCreated, nil, build, nil
}

func (handler *Handler) get(url *url.URL, header http.Header) (int, http.Header, []Build, error) {
	handler.buildsMutex.RLock()

	builds := make([]Build, len(handler.builds))

	i := 0
	for _, build := range handler.builds {
		builds[i] = *build
		i++
	}

	handler.buildsMutex.RUnlock()

	sort.Sort(sort.Reverse(ByCreatedAt(builds)))

	return http.StatusOK, nil, builds, nil
}

func (handler *Handler) validateBuild(build *Build) error {
	if build.Image == "" {
		return errors.New("missing build image")
	}

	return nil
}
