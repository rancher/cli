package hosts

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/rancher/rke/k8s"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"golang.org/x/crypto/ssh"
)

const (
	DockerDialerTimeout = 30
)

type DialerFactory func(h *Host) (func(network, address string) (net.Conn, error), error)

type dialer struct {
	signer          ssh.Signer
	sshKeyString    string
	sshAddress      string
	username        string
	netConn         string
	dockerSocket    string
	useSSHAgentAuth bool
	bastionDialer   *dialer
}

func newDialer(h *Host, kind string) (*dialer, error) {
	// Check for Bastion host connection
	var bastionDialer *dialer
	if len(h.BastionHost.Address) > 0 {
		bastionDialer = &dialer{
			sshAddress:      fmt.Sprintf("%s:%s", h.BastionHost.Address, h.BastionHost.Port),
			username:        h.BastionHost.User,
			sshKeyString:    h.BastionHost.SSHKey,
			netConn:         "tcp",
			useSSHAgentAuth: h.SSHAgentAuth,
		}
		if bastionDialer.sshKeyString == "" {
			bastionDialer.sshKeyString = privateKeyPath(h.BastionHost.SSHKeyPath)
		}
	}

	dialer := &dialer{
		sshAddress:      fmt.Sprintf("%s:%s", h.Address, h.Port),
		username:        h.User,
		dockerSocket:    h.DockerSocket,
		sshKeyString:    h.SSHKey,
		netConn:         "unix",
		useSSHAgentAuth: h.SSHAgentAuth,
		bastionDialer:   bastionDialer,
	}

	if dialer.sshKeyString == "" {
		dialer.sshKeyString = privateKeyPath(h.SSHKeyPath)
	}

	switch kind {
	case "network", "health":
		dialer.netConn = "tcp"
	}

	if len(dialer.dockerSocket) == 0 {
		dialer.dockerSocket = "/var/run/docker.sock"
	}

	return dialer, nil
}

func SSHFactory(h *Host) (func(network, address string) (net.Conn, error), error) {
	dialer, err := newDialer(h, "docker")
	return dialer.Dial, err
}

func LocalConnFactory(h *Host) (func(network, address string) (net.Conn, error), error) {
	dialer, err := newDialer(h, "network")
	return dialer.Dial, err
}

func (d *dialer) DialDocker(network, addr string) (net.Conn, error) {
	return d.Dial(network, addr)
}

func (d *dialer) DialLocalConn(network, addr string) (net.Conn, error) {
	return d.Dial(network, addr)
}

func (d *dialer) Dial(network, addr string) (net.Conn, error) {
	var conn *ssh.Client
	var err error
	if d.bastionDialer != nil {
		conn, err = d.getBastionHostTunnelConn()
	} else {
		conn, err = d.getSSHTunnelConnection()
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to dial ssh using address [%s]: %v", d.sshAddress, err)
	}

	// Docker Socket....
	if d.netConn == "unix" {
		addr = d.dockerSocket
		network = d.netConn
	}

	remote, err := conn.Dial(network, addr)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial to %s: %v", addr, err)
	}
	return remote, err
}

func (d *dialer) getSSHTunnelConnection() (*ssh.Client, error) {
	cfg, err := getSSHConfig(d.username, d.sshKeyString, d.useSSHAgentAuth)
	if err != nil {
		return nil, fmt.Errorf("Error configuring SSH: %v", err)
	}
	// Establish connection with SSH server
	return ssh.Dial("tcp", d.sshAddress, cfg)
}

func (h *Host) newHTTPClient(dialerFactory DialerFactory) (*http.Client, error) {
	factory := dialerFactory
	if factory == nil {
		factory = SSHFactory
	}

	dialer, err := factory(h)
	if err != nil {
		return nil, err
	}
	dockerDialerTimeout := time.Second * DockerDialerTimeout
	return &http.Client{
		Transport: &http.Transport{
			Dial:                  dialer,
			TLSHandshakeTimeout:   dockerDialerTimeout,
			IdleConnTimeout:       dockerDialerTimeout,
			ResponseHeaderTimeout: dockerDialerTimeout,
		},
	}, nil
}

func (d *dialer) getBastionHostTunnelConn() (*ssh.Client, error) {
	bastionCfg, err := getSSHConfig(d.bastionDialer.username, d.bastionDialer.sshKeyString, d.bastionDialer.useSSHAgentAuth)
	if err != nil {
		return nil, fmt.Errorf("Error configuring SSH for bastion host [%s]: %v", d.bastionDialer.sshAddress, err)
	}
	bastionClient, err := ssh.Dial("tcp", d.bastionDialer.sshAddress, bastionCfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to the bastion host [%s]: %v", d.bastionDialer.sshAddress, err)
	}
	conn, err := bastionClient.Dial(d.bastionDialer.netConn, d.sshAddress)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to the host [%s]: %v", d.sshAddress, err)
	}
	cfg, err := getSSHConfig(d.username, d.sshKeyString, d.useSSHAgentAuth)
	if err != nil {
		return nil, fmt.Errorf("Error configuring SSH for host [%s]: %v", d.sshAddress, err)
	}
	newClientConn, channels, sshRequest, err := ssh.NewClientConn(conn, d.sshAddress, cfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to establish new ssh client conn [%s]: %v", d.sshAddress, err)
	}
	return ssh.NewClient(newClientConn, channels, sshRequest), nil
}

func BastionHostWrapTransport(bastionHost v3.BastionHost) k8s.WrapTransport {

	bastionDialer := &dialer{
		sshAddress:      fmt.Sprintf("%s:%s", bastionHost.Address, bastionHost.Port),
		username:        bastionHost.User,
		sshKeyString:    bastionHost.SSHKey,
		netConn:         "tcp",
		useSSHAgentAuth: bastionHost.SSHAgentAuth,
	}

	if bastionDialer.sshKeyString == "" {
		bastionDialer.sshKeyString = privateKeyPath(bastionHost.SSHKeyPath)
	}
	return func(rt http.RoundTripper) http.RoundTripper {
		if ht, ok := rt.(*http.Transport); ok {
			ht.DialContext = nil
			ht.DialTLS = nil
			ht.Dial = bastionDialer.Dial
		}
		return rt
	}
}
