package logbuffer

import (
	"sync"

	"github.com/gorilla/websocket"
)

type LogBuffer struct {
	content      []byte
	contentMutex *sync.RWMutex

	sinks []*websocket.Conn
}

func NewLogBuffer() *LogBuffer {
	return &LogBuffer{
		contentMutex: new(sync.RWMutex),
	}
}

func (buffer *LogBuffer) Write(data []byte) (int, error) {
	buffer.contentMutex.Lock()

	buffer.content = append(buffer.content, data...)

	newSinks := []*websocket.Conn{}
	for _, sink := range buffer.sinks {
		err := sink.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			continue
		}

		newSinks = append(newSinks, sink)
	}

	buffer.sinks = newSinks

	buffer.contentMutex.Unlock()

	return len(data), nil
}

func (buffer *LogBuffer) Attach(conn *websocket.Conn) {
	buffer.contentMutex.Lock()

	buffer.sinks = append(buffer.sinks, conn)

	conn.WriteMessage(websocket.BinaryMessage, buffer.content)

	buffer.contentMutex.Unlock()
}
