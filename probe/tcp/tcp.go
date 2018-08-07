package tcp

import (
	"errors"
	"net"
	"time"
)

type TcpProbe struct {
	server string
	port   string
	addr   string
}

// New returns a new tcpProbe that implements the Prober interface
func New(server, port string) (*TcpProbe, error) {
	if server == "" || port == "" {
		return nil, errors.New("need a server and port to create tcp probe")
	}
	p := TcpProbe{
		server: server,
		port:   port,
		addr:   server + port,
	}
	return &p, nil
}

func (p *TcpProbe) Probe() error {
	_, err := net.DialTimeout("tcp", p.addr, 5*time.Second)
	if err != nil {
		return err
	}
	return nil
}
