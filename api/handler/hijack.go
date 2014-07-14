package handler

import (
	"io"

	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/pivotal-golang/lager"
)

func (handler *Handler) HijackBuild(w http.ResponseWriter, r *http.Request) {
	guid := r.FormValue(":guid")

	handler.buildsMutex.RLock()
	build, found := handler.builds[guid]
	handler.buildsMutex.RUnlock()

	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log := handler.logger.Session("hijack", lager.Data{
		"build": build,
	})

	log.Info("hijacking")

	hijackURL, err := url.Parse(build.HijackURL)
	if err != nil {
		log.Error("failed-to-parse-url", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	conn, err := net.Dial("tcp", hijackURL.Host)
	if err != nil {
		log.Error("failed-to-dial-turbine", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	client := httputil.NewClientConn(conn, nil)

	req, err := http.NewRequest(r.Method, build.HijackURL, r.Body)
	if err != nil {
		log.Error("failed-to-create-request", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error("failed-to-hijack", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.StatusCode)

	sconn, sbr, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Error("failed-to-hijack", err)
		return
	}

	cconn, cbr := client.Hijack()

	go io.Copy(cconn, sbr)

	io.Copy(sconn, cbr)
}
