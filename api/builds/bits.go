package builds

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func (handler *Handler) PostBits(w http.ResponseWriter, r *http.Request) {
	guid := mux.Vars(r)["guid"]

	handler.buildsMutex.RLock()
	build, found := handler.builds[guid]
	handler.buildsMutex.RUnlock()

	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log.Println("triggering", build.Guid)

	buf := new(bytes.Buffer)

	proleBuild := ProleBuild{
		Guid: build.Guid,

		LogsURL: "ws://" + handler.peerAddr + "/builds/" + build.Guid + "/log/input",

		Image:  build.Image,
		Script: build.Script,

		Source: ProleBuildSource{
			Type: "raw",
			URI:  "http://" + handler.peerAddr + "/builds/" + build.Guid + "/bits",
			Path: build.Path,
		},

		Callback: "http://" + handler.peerAddr + "/builds/" + build.Guid + "/result",

		Parameters: build.Environment,
	}

	err := json.NewEncoder(buf).Encode(proleBuild)
	if err != nil {
		panic(err)
	}

	defer r.Body.Close()

	res, err := http.Post(handler.proleURL+"/builds", "application/json", buf)
	if err != nil {
		log.Println("error triggering build:", err)
		panic(err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	res.Body.Close()

	if res.StatusCode == http.StatusCreated {
		build.servingBits.Add(1)

		w.WriteHeader(http.StatusCreated)

		build.bits <- r

		build.servingBits.Wait()
	} else {
		log.Println("prole failed:")
		res.Write(os.Stderr)
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

func (handler *Handler) GetBits(w http.ResponseWriter, r *http.Request) {
	guid := mux.Vars(r)["guid"]

	handler.buildsMutex.RLock()
	build, found := handler.builds[guid]
	handler.buildsMutex.RUnlock()

	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var bits *http.Request

	select {
	case bits = <-build.bits:
	case <-time.After(time.Second):
		w.WriteHeader(404)
		return
	}

	log.Println("serving bits for", build.Guid)

	defer build.servingBits.Done()

	w.Header().Set("Content-Type", bits.Header.Get("Content-Type"))

	w.WriteHeader(200)

	_, err := io.Copy(w, bits.Body)
	if err != nil {
		log.Println("streaming bits failed:", err)
	}
}
