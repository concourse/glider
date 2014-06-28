package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/nu7hatch/gouuid"
	"github.com/pivotal-golang/lager"

	"github.com/concourse/glider/api/builds"
	"github.com/concourse/logbuffer"
)

func (handler *Handler) CreateBuild(w http.ResponseWriter, r *http.Request) {
	var build builds.Build
	err := json.NewDecoder(r.Body).Decode(&build)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = handler.validateBuild(build)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	build.Guid = uuid.String()
	build.CreatedAt = time.Now()

	log := handler.logger.Session("create", lager.Data{
		"build": build,
	})

	log.Info("register")

	handler.bitsMutex.Lock()
	handler.bits[build.Guid] = BitsSession{
		bits:        make(chan *http.Request, 1),
		servingBits: &sync.WaitGroup{},
	}
	handler.bitsMutex.Unlock()

	handler.logsMutex.Lock()
	handler.logs[build.Guid] = logbuffer.NewLogBuffer()
	handler.logsMutex.Unlock()

	handler.buildsMutex.Lock()
	handler.builds[build.Guid] = &build
	handler.buildsMutex.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(build)
}

func (handler *Handler) GetBuild(w http.ResponseWriter, r *http.Request) {
	handler.buildsMutex.RLock()

	builds := make([]builds.Build, len(handler.builds))

	i := 0
	for _, build := range handler.builds {
		builds[i] = *build
		i++
	}

	handler.buildsMutex.RUnlock()

	sort.Sort(sort.Reverse(ByCreatedAt(builds)))

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(builds)
}

func (handler *Handler) validateBuild(build builds.Build) error {
	if build.Image == "" {
		return errors.New("missing build image")
	}

	return nil
}
