package api

import (
	"net/http"

	"code.google.com/p/go.net/websocket"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"

	"github.com/concourse/glider/api/handler"
	"github.com/concourse/glider/routes"
)

func New(logger lager.Logger, peerAddr, turbineURL string) (http.Handler, error) {
	builds := handler.NewHandler(logger, peerAddr, turbineURL)

	handlers := map[string]http.Handler{
		routes.CreateBuild: http.HandlerFunc(builds.CreateBuild),
		routes.GetBuild:    http.HandlerFunc(builds.GetBuild),
		routes.HijackBuild: http.HandlerFunc(builds.HijackBuild),

		routes.UploadBits:   http.HandlerFunc(builds.UploadBits),
		routes.DownloadBits: http.HandlerFunc(builds.DownloadBits),

		routes.SetResult: http.HandlerFunc(builds.SetResult),
		routes.GetResult: http.HandlerFunc(builds.GetResult),

		routes.LogInput:  websocket.Server{Handler: builds.LogInput},
		routes.LogOutput: websocket.Server{Handler: builds.LogOutput},
	}

	return rata.NewRouter(routes.Routes, handlers)
}
