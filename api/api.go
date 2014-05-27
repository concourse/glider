package api

import (
	"net/http"

	"code.google.com/p/go.net/websocket"
	"github.com/tedsuo/router"

	"github.com/winston-ci/redgreen/api/handler"
	"github.com/winston-ci/redgreen/routes"
)

func New(peerAddr, proleURL string) (http.Handler, error) {
	builds := handler.NewHandler(peerAddr, proleURL)

	handlers := map[string]http.Handler{
		routes.CreateBuild: http.HandlerFunc(builds.CreateBuild),
		routes.GetBuild:    http.HandlerFunc(builds.GetBuild),

		routes.UploadBits:   http.HandlerFunc(builds.UploadBits),
		routes.DownloadBits: http.HandlerFunc(builds.DownloadBits),

		routes.SetResult: http.HandlerFunc(builds.SetResult),
		routes.GetResult: http.HandlerFunc(builds.GetResult),

		routes.LogInput:  websocket.Server{Handler: builds.LogInput},
		routes.LogOutput: websocket.Server{Handler: builds.LogOutput},
	}

	return router.NewRouter(routes.Routes, handlers)
}
