package dockerapiproxy

import (
	"errors"
	"net"
	"time"

	"github.com/gorilla/websocket"
	rancher "github.com/rancher/go-rancher/v3"
)

type Dialer func(network, addr string) (net.Conn, error)

type dialer struct {
	p    *Proxy
	host *rancher.Host
}

func NewDialer(client *rancher.RancherClient, host string) (Dialer, error) {
	d := &dialer{
		p: &Proxy{
			client: client,
			host:   host,
		},
	}
	rancherHost, err := d.p.getHost()
	if err != nil {
		return nil, err
	}
	d.host = rancherHost
	return d.Dial, nil
}

func (d *dialer) Dial(network, addr string) (net.Conn, error) {
	if network != "tcp" {
		return nil, errors.New("Only tcp network is allowed")
	}

	conn, err := d.p.openConnection(d.host)
	return &wsConn{
		Conn: conn,
		conn: &WebSocketIo{conn},
	}, err
}

type wsConn struct {
	*websocket.Conn
	conn *WebSocketIo
	temp []byte
}

func (w *wsConn) Read(buf []byte) (int, error) {
	var err error
	if len(w.temp) > 0 {
		return w.readFromTemp(buf), nil
	}

	w.temp, err = w.conn.Read()
	return w.readFromTemp(buf), err
}

func (w *wsConn) Write(p []byte) (n int, err error) {
	return w.conn.Write(p)
}

func (w *wsConn) SetDeadline(t time.Time) error {
	if err := w.SetReadDeadline(t); err != nil {
		return err
	}
	return w.SetWriteDeadline(t)
}

func (w *wsConn) readFromTemp(buf []byte) int {
	n := copy(buf, w.temp)
	w.temp = w.temp[n:]
	return n
}
