package io

import (
	"github.com/gorilla/websocket"
)

type WebSocketReadWriteCloser struct {
	Conn *websocket.Conn
}

func (w *WebSocketReadWriteCloser) Read(p []byte) (int, error) {
	_, data, err := w.Conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	return copy(p, data), nil
}

func (w *WebSocketReadWriteCloser) Write(p []byte) (int, error) {
	err := w.Conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *WebSocketReadWriteCloser) Close() error {
	return w.Conn.Close()
}
