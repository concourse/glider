package handler

import (
	"io"

	"github.com/pivotal-golang/lager"

	"code.google.com/p/go.net/websocket"
)

func (handler *Handler) LogInput(conn *websocket.Conn) {
	guid := conn.Request().FormValue(":guid")

	handler.logsMutex.RLock()
	logBuffer, found := handler.logs[guid]
	handler.logsMutex.RUnlock()

	if !found {
		return
	}

	defer logBuffer.Close()

	log := handler.logger.Session("log-in", lager.Data{
		"guid": guid,
	})

	_, err := io.Copy(logBuffer, conn)
	if err != nil {
		log.Error("failed-to-stream", err)
		return
	}
}

func (handler *Handler) LogOutput(conn *websocket.Conn) {
	guid := conn.Request().FormValue(":guid")

	handler.logsMutex.RLock()
	logBuffer, found := handler.logs[guid]
	handler.logsMutex.RUnlock()

	if !found {
		return
	}

	logBuffer.Attach(conn)
}
