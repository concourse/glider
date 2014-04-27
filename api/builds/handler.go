package builds

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/nu7hatch/gouuid"
	"github.com/rcrowley/go-tigertonic"
)

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

func (handler *Handler) PostHandler() http.Handler {
	return tigertonic.Marshaled(handler.post)
}

func (handler *Handler) GetHandler() http.Handler {
	return tigertonic.Marshaled(handler.get)
}

func (handler *Handler) PostBitsHandler() http.Handler {
	return http.HandlerFunc(handler.postBits)
}

func (handler *Handler) GetBitsHandler() http.Handler {
	return http.HandlerFunc(handler.getBits)
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
	build.bits = make(chan *http.Request, 1)

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

func (handler *Handler) postBits(w http.ResponseWriter, req *http.Request) {
	guid := req.URL.Query().Get("guid")

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

		LogConfig: models.LogConfig{
			Guid:       build.Guid,
			SourceName: "BLD",
		},

		Image:  build.Image,
		Script: build.Script,

		Source: ProleBuildSource{
			Type: "raw",
			URI:  "http://" + handler.peerAddr + "/builds/" + build.Guid + "/bits",
		},

		Callback: "http://" + handler.peerAddr + "/builds/" + build.Guid + "/result",

		Parameters: build.Environment,
	}

	err := json.NewEncoder(buf).Encode(proleBuild)
	if err != nil {
		panic(err)
	}

	build.servingBits.Add(1)

	res, err := http.Post(handler.proleURL+"/builds", "application/json", buf)
	if err != nil {
		log.Println("error triggering build:", err)
		panic(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else if res.StatusCode == http.StatusCreated {
		w.WriteHeader(http.StatusCreated)

		build.bits <- req

		build.servingBits.Wait()
	} else {
		log.Println("prole failed:")
		res.Write(os.Stderr)
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

func (handler *Handler) getBits(w http.ResponseWriter, req *http.Request) {
	guid := req.URL.Query().Get("guid")

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

	defer bits.Body.Close()
	defer build.servingBits.Done()

	w.Header().Set("Content-Type", bits.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", bits.Header.Get("Content-Length"))

	w.WriteHeader(200)

	_, err := io.Copy(w, bits.Body)
	if err != nil {
		log.Println("streaming bits failed:", err)
	}
}

func (handler *Handler) validateBuild(build *Build) error {
	if build.Image == "" {
		return errors.New("missing build image")
	}

	return nil
}
