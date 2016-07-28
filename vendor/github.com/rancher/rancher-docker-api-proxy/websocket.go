package dockerapiproxy

import (
	"encoding/base64"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
)

type WebSocketIo struct {
	Conn *websocket.Conn
}

func (w *WebSocketIo) Read() ([]byte, error) {
	_, bytes, err := w.Conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	bytes, err = base64.StdEncoding.DecodeString(string(bytes))
	logrus.Debugf("Websocket Read: %d: %s", len(bytes), string(bytes))
	return bytes, err
}

func (w *WebSocketIo) Write(buf []byte) (int, error) {
	logrus.Debugf("Websocket Writer: %d: %s", len(buf), string(buf))
	str := base64.StdEncoding.EncodeToString(buf)
	err := w.Conn.WriteMessage(websocket.TextMessage, []byte(str))
	return len(buf), err
}
