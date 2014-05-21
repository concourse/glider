package handler

import (
	"io"
	"log"

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

	_, err := io.Copy(logBuffer, conn)
	if err != nil {
		log.Println("error reading message:", err)
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
