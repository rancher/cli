// +build !windows

package dockerapiproxy

import "net"

func getListener(proto, address string) (net.Listener, error) {
	return net.Listen(proto, address)
}
