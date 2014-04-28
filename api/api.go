package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/winston-ci/redgreen/api/builds"
)

func New(logger *log.Logger, peerAddr, proleURL string) http.Handler {
	router := mux.NewRouter()

	builds := builds.NewHandler(peerAddr, proleURL)

	router.Methods("POST").Path("/builds").HandlerFunc(builds.Post)
	router.Methods("GET").Path("/builds").HandlerFunc(builds.Get)

	buildRouter := router.PathPrefix("/builds/{guid}").Subrouter()

	buildRouter.Methods("POST").Path("/bits").HandlerFunc(builds.PostBits)
	buildRouter.Methods("GET").Path("/bits").HandlerFunc(builds.GetBits)

	buildRouter.Methods("PUT").Path("/result").HandlerFunc(builds.PutResult)
	buildRouter.Methods("GET").Path("/result").HandlerFunc(builds.GetResult)

	buildRouter.Path("/log/input").HandlerFunc(builds.LogInput)
	buildRouter.Path("/log/output").HandlerFunc(builds.LogOutput)

	return router
}
