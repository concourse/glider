package api

import (
	"log"
	"net/http"

	"github.com/rcrowley/go-tigertonic"
	"github.com/winston-ci/redgreen/api/builds"
)

func New(logger *log.Logger, peerAddr, proleURL string) http.Handler {
	mux := tigertonic.NewTrieServeMux()

	builds := builds.NewHandler(peerAddr, proleURL)

	mux.Handle("POST", "/builds", logged(logger, builds.PostHandler()))
	mux.Handle("GET", "/builds", logged(logger, builds.GetHandler()))

	mux.Handle("POST", "/builds/{guid}/bits", logged(logger, builds.PostBitsHandler()))
	mux.Handle("GET", "/builds/{guid}/bits", logged(logger, builds.GetBitsHandler()))

	mux.Handle("PUT", "/builds/{guid}/result", logged(logger, builds.PutResultHandler()))
	mux.Handle("GET", "/builds/{guid}/result", logged(logger, builds.GetResultHandler()))

	return mux
}

func logged(logger *log.Logger, handler http.Handler) http.Handler {
	logged := tigertonic.Logged(handler, nil)
	logged.Logger = logger
	return logged
}
