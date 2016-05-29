package dockerapiproxy

import "net"

type SocketIo struct {
	Conn net.Conn
}

func (s *SocketIo) Read() ([]byte, error) {
	buf := make([]byte, 8192)
	c, err := s.Conn.Read(buf)
	return buf[:c], err
}

func (s *SocketIo) Write(buf []byte) (int, error) {
	return s.Conn.Write(buf)
}
