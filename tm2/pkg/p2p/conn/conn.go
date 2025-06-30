package conn

import (
	"net"
	"time"
)

// pipe wraps the networking conn interface
type pipe struct {
	net.Conn
}

func (p *pipe) SetDeadline(_ time.Time) error {
	return nil
}

func NetPipe() (net.Conn, net.Conn) {
	p1, p2 := net.Pipe()
	return &pipe{p1}, &pipe{p2}
}

var _ net.Conn = (*pipe)(nil)
