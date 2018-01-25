// +build windows

package dockerapiproxy

import (
	"net"

	"github.com/Microsoft/go-winio"
)

func getListener(proto, address string) (net.Listener, error) {
	npipeConfig := &winio.PipeConfig{
		SecurityDescriptor: "D:P(A;;GA;;;BA)(A;;GA;;;SY)",
		MessageMode:        true,  // Use message mode so that CloseWrite() is supported
		InputBufferSize:    65536, // Use 64KB buffers to improve performance
		OutputBufferSize:   65536,
	}
	return winio.ListenPipe(address, npipeConfig)
}
