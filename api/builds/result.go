package builds

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (handler *Handler) PutResult(w http.ResponseWriter, r *http.Request) {
	guid := mux.Vars(r)["guid"]

	handler.buildsMutex.RLock()
	build, found := handler.builds[guid]
	handler.buildsMutex.RUnlock()

	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var result BuildResult
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println("updating result", build.Guid, result)

	handler.buildsMutex.Lock()
	build.Status = result.Status
	build.logBuffer.Close()
	handler.buildsMutex.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (handler *Handler) GetResult(w http.ResponseWriter, r *http.Request) {
	guid := mux.Vars(r)["guid"]

	handler.buildsMutex.RLock()
	build, found := handler.builds[guid]
	handler.buildsMutex.RUnlock()

	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(BuildResult{build.Status})
}
