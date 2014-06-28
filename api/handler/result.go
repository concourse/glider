package handler

import (
	"encoding/json"
	"net/http"

	"github.com/concourse/glider/api/builds"
	"github.com/pivotal-golang/lager"
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

	log := handler.logger.Session("set-result", lager.Data{
		"build": build,
	})

	var result builds.BuildResult
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Info("update", lager.Data{
		"result": result,
	})

	handler.buildsMutex.Lock()
	build.Status = result.Status
	handler.buildsMutex.Unlock()

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
