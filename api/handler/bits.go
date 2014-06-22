package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/concourse/turbine/api/builds"
)

func (handler *Handler) UploadBits(w http.ResponseWriter, r *http.Request) {
	guid := r.FormValue(":guid")

	handler.buildsMutex.RLock()
	build, found := handler.builds[guid]
	handler.buildsMutex.RUnlock()

	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log.Println("triggering", build.Guid)

	buf := new(bytes.Buffer)

	turbineBuild := builds.Build{
		Guid: build.Guid,

		LogsURL: "ws://" + handler.peerAddr + "/builds/" + build.Guid + "/log/input",

		Privileged: true,

		Config: builds.Config{
			Image:  build.Image,
			Script: build.Script,
			Env:    build.Env,
		},

		Inputs: []builds.Input{
			{
				Type: "raw",

				Source: builds.Source{
					"uri": "http://" + handler.peerAddr + "/builds/" + build.Guid + "/bits",
				},

				DestinationPath: build.Path,
			},
		},

		Callback: "http://" + handler.peerAddr + "/builds/" + build.Guid + "/result",
	}

	err := json.NewEncoder(buf).Encode(turbineBuild)
	if err != nil {
		panic(err)
	}

	defer r.Body.Close()

	res, err := http.Post(handler.turbineURL+"/builds", "application/json", buf)
	if err != nil {
		log.Println("error triggering build:", err)
		panic(err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	res.Body.Close()

	if res.StatusCode == http.StatusCreated {
		w.WriteHeader(http.StatusCreated)

		handler.bitsMutex.RLock()
		session := handler.bits[guid]
		handler.bitsMutex.RUnlock()

		session.servingBits.Add(1)
		session.bits <- r
		session.servingBits.Wait()
	} else {
		log.Println("turbine failed:")
		res.Write(os.Stderr)
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

func (handler *Handler) DownloadBits(w http.ResponseWriter, r *http.Request) {
	guid := r.FormValue(":guid")

	handler.bitsMutex.RLock()
	session, found := handler.bits[guid]
	handler.bitsMutex.RUnlock()

	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var bits *http.Request

	select {
	case bits = <-session.bits:
	case <-time.After(time.Second):
		w.WriteHeader(404)
		return
	}

	log.Println("serving bits for", guid)

	defer session.servingBits.Done()

	w.Header().Set("Content-Type", bits.Header.Get("Content-Type"))

	w.WriteHeader(200)

	_, err := io.Copy(w, bits.Body)
	if err != nil {
		log.Println("streaming bits failed:", err)
	}
}
