package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/winston-ci/redgreen/api/builds"
)

func (handler *Handler) SetResult(w http.ResponseWriter, r *http.Request) {
	guid := r.FormValue(":guid")

	handler.buildsMutex.RLock()
	build, found := handler.builds[guid]
	handler.buildsMutex.RUnlock()

	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var result builds.BuildResult
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println("updating result", build.Guid, result)

	handler.buildsMutex.Lock()
	build.Status = result.Status
	handler.buildsMutex.Unlock()

	handler.logsMutex.Lock()
	handler.logs[build.Guid].Close()
	handler.logsMutex.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (handler *Handler) GetResult(w http.ResponseWriter, r *http.Request) {
	guid := r.FormValue(":guid")

	handler.buildsMutex.RLock()
	build, found := handler.builds[guid]
	handler.buildsMutex.RUnlock()

	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(builds.BuildResult{build.Status})
}
