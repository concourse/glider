package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pivotal-golang/lager"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool {
		return true
	},
}

func (handler *Handler) LogInput(w http.ResponseWriter, r *http.Request) {
	guid := r.FormValue(":guid")

	log := handler.logger.Session("log-in", lager.Data{
		"guid": guid,
	})

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("failed-to-upgrade", err)
		return
	}

	handler.logsMutex.RLock()
	logBuffer, found := handler.logs[guid]
	handler.logsMutex.RUnlock()

	if !found {
		return
	}

	defer logBuffer.Close()

	for {
		var msg *json.RawMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		logBuffer.WriteMessage(msg)
	}
}

func (handler *Handler) LogOutput(w http.ResponseWriter, r *http.Request) {
	guid := r.FormValue(":guid")

	log := handler.logger.Session("log-out", lager.Data{
		"guid": guid,
	})

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("failed-to-upgrade", err)
		return
	}

	handler.logsMutex.RLock()
	logBuffer, found := handler.logs[guid]
	handler.logsMutex.RUnlock()

	if !found {
		return
	}

	logBuffer.Attach(conn)
}
