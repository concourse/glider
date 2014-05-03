package builds

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/nu7hatch/gouuid"
	"github.com/winston-ci/redgreen/logbuffer"
)

func (handler *Handler) CreateBuild(w http.ResponseWriter, r *http.Request) {
	var build Build
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
	build.bits = make(chan *http.Request, 1)
	build.servingBits = &sync.WaitGroup{}
	build.logBuffer = logbuffer.NewLogBuffer()

	log.Println("registering", build.Guid)

	handler.buildsMutex.Lock()
	handler.builds[build.Guid] = &build
	handler.buildsMutex.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(build)
}

func (handler *Handler) GetBuild(w http.ResponseWriter, r *http.Request) {
	handler.buildsMutex.RLock()

	builds := make([]Build, len(handler.builds))

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

func (handler *Handler) validateBuild(build Build) error {
	if build.Image == "" {
		return errors.New("missing build image")
	}

	return nil
}
