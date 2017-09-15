package dockerapiproxy

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"github.com/rancher/go-rancher/hostaccess"
	rancher "github.com/rancher/go-rancher/v3"
)

type IoOps interface {
	io.Writer
	Read() ([]byte, error)
}

type Proxy struct {
	client       *rancher.RancherClient
	host, listen string
	TlsConfig    *tls.Config
	listener     net.Listener
	rancherHost  *rancher.Host
}

func NewProxy(client *rancher.RancherClient, host, listen string) *Proxy {
	return &Proxy{
		client: client,
		host:   host,
		listen: listen,
	}
}

func (p *Proxy) Close() error {
	return p.listener.Close()
}

func (p *Proxy) getSocket(url string) (net.Listener, error) {
	proto := "tcp"
	address := url

	parts := strings.SplitN(url, "://", 2)
	if len(parts) == 2 {
		proto = parts[0]
		address = parts[1]
	}

	if proto == "unix" {
		os.Remove(address)
	}

	l, err := net.Listen(proto, address)
	if err != nil {
		return nil, err
	}

	if proto == "tcp" && p.TlsConfig != nil {
		p.TlsConfig.NextProtos = []string{"http/1.1"}
		l = tls.NewListener(l, p.TlsConfig)
	}

	return l, err
}

func (p *Proxy) Listen() error {
	host, err := p.getHost()
	if err != nil {
		return err
	}

	os.Remove(p.listen)

	l, err := p.getSocket(p.listen)
	if err != nil {
		return err
	}

	p.listener = l
	p.rancherHost = host
	return nil
}

func (p *Proxy) Serve() error {
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			return err
		}

		logrus.Debug("New connection")
		go p.handle(p.rancherHost, conn)
	}
}

func (p *Proxy) ListenAndServe() error {
	if err := p.Listen(); err != nil {
		return err
	}
	return p.Serve()
}

func (p *Proxy) handle(host *rancher.Host, client net.Conn) {
	if err := p.handleError(host, client); err != nil {
		logrus.Errorf("Failed to handle connection: %v", err)
	}
}

func (p *Proxy) openConnection(host *rancher.Host) (*websocket.Conn, error) {
	hostAccessClient := hostaccess.RancherWebsocketClient(*p.client)
	return hostAccessClient.GetHostAccess(host.Resource, "dockersocket", nil)
}

func (p *Proxy) handleError(host *rancher.Host, conn net.Conn) error {
	defer conn.Close()

	websocket, err := p.openConnection(host)
	if err != nil {
		return err
	}

	server := &WebSocketIo{Conn: websocket}
	client := &SocketIo{Conn: conn}

	wg := sync.WaitGroup{}
	wg.Add(2)

	abort := func() {
		wg.Done()
		conn.Close()
		websocket.Close()
	}

	go func() {
		defer abort()
		p.copyLoop(client, server)
	}()

	go func() {
		defer abort()
		p.copyLoop(server, client)
	}()

	wg.Wait()

	logrus.Debugf("Disconnecting")

	return nil
}

func (p *Proxy) copyLoop(from, to IoOps) error {
	con := true

	for con {
		buf, err := from.Read()
		if err != nil {
			return err
		}
		logrus.Debugf("Read %d bytes", len(buf))
		if _, err := to.Write(buf); err != nil {
			return err
		}
		logrus.Debugf("Wrote %d bytes", len(buf))
	}

	return nil
}

func (p *Proxy) getHost() (*rancher.Host, error) {
	host, err := p.client.Host.ById(p.host)
	if err != nil {
		return nil, err
	}

	if host != nil {
		return host, nil
	}

	hosts, err := p.client.Host.List(&rancher.ListOpts{
		Filters: map[string]interface{}{
			"uuid": p.host,
		},
	})
	if err != nil {
		return nil, err
	}

	if len(hosts.Data) == 0 {
		hosts, err = p.client.Host.List(&rancher.ListOpts{
			Filters: map[string]interface{}{
				"name": p.host,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	if len(hosts.Data) == 0 {
		hosts, err = p.client.Host.List(&rancher.ListOpts{
			Filters: map[string]interface{}{
				"hostname": p.host,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	if len(hosts.Data) == 0 {
		return nil, fmt.Errorf("Failed to find host: %s", p.host)
	}

	return &hosts.Data[0], nil
}
