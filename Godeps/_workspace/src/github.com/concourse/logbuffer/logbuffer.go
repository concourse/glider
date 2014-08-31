package logbuffer

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/gorilla/websocket"
)

type LogBuffer struct {
	content      []*json.RawMessage
	contentMutex *sync.RWMutex

	sinks []*websocket.Conn

	closed        bool
	waitForClosed chan struct{}
}

func NewLogBuffer() *LogBuffer {
	return &LogBuffer{
		contentMutex:  new(sync.RWMutex),
		waitForClosed: make(chan struct{}),
	}
}

func (buffer *LogBuffer) WriteMessage(msg *json.RawMessage) error {
	buffer.contentMutex.Lock()

	buffer.content = append(buffer.content, msg)

	newSinks := []*websocket.Conn{}
	for _, sink := range buffer.sinks {
		err := sink.WriteJSON(msg)
		if err != nil {
			continue
		}

		newSinks = append(newSinks, sink)
	}

	buffer.sinks = newSinks

	buffer.contentMutex.Unlock()

	return nil
}

func (buffer *LogBuffer) Attach(sink *websocket.Conn) {
	buffer.contentMutex.Lock()

	for _, msg := range buffer.content {
		err := sink.WriteJSON(msg)
		if err != nil {
			return
		}
	}

	if buffer.closed {
		sink.Close()
	} else {
		buffer.sinks = append(buffer.sinks, sink)
	}

	buffer.contentMutex.Unlock()

	<-buffer.waitForClosed
}

func (buffer *LogBuffer) Close() error {
	buffer.contentMutex.Lock()
	defer buffer.contentMutex.Unlock()

	if buffer.closed {
		return errors.New("close twice")
	}

	for _, sink := range buffer.sinks {
		sink.Close()
	}

	buffer.closed = true
	buffer.sinks = nil

	close(buffer.waitForClosed)

	return nil
}
