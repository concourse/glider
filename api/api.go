package api

import (
	"log"
	"net/http"

	"github.com/tedsuo/router"

	"github.com/winston-ci/redgreen/api/handler"
	"github.com/winston-ci/redgreen/routes"
)

func New(logger *log.Logger, peerAddr, proleURL string) (http.Handler, error) {
	builds := handler.NewHandler(peerAddr, proleURL)

	handlers := map[string]http.Handler{
		routes.CreateBuild: http.HandlerFunc(builds.CreateBuild),
		routes.GetBuild:    http.HandlerFunc(builds.GetBuild),

		routes.UploadBits:   http.HandlerFunc(builds.UploadBits),
		routes.DownloadBits: http.HandlerFunc(builds.DownloadBits),

		routes.SetResult: http.HandlerFunc(builds.SetResult),
		routes.GetResult: http.HandlerFunc(builds.GetResult),

		routes.LogInput:  http.HandlerFunc(builds.LogInput),
		routes.LogOutput: http.HandlerFunc(builds.LogOutput),
	}

	return router.NewRouter(routes.Routes, handlers)
}
