package api

import (
	"log"
	"net/http"

	"github.com/rcrowley/go-tigertonic"
	"github.com/winston-ci/redgreen/api/builds"
)

func New(logger *log.Logger, proleURL string) http.Handler {
	mux := tigertonic.NewTrieServeMux()

	builds := builds.NewHandler(proleURL)

	mux.Handle("POST", "/builds", logged(logger, builds.PostHandler()))
	mux.Handle("GET", "/builds", logged(logger, builds.GetHandler()))

	return mux
}

func logged(logger *log.Logger, handler http.Handler) http.Handler {
	logged := tigertonic.Logged(handler, nil)
	logged.Logger = logger
	return logged
}
