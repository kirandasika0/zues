package probe

import "zues/probe/tcp"

// Prober is a lightweight TCP probe that runs before every stress test
type Prober interface {
	// Probe is a method that probes at the given URL
	Probe() error
}

// NewTCPProbe returns a struct that implements the prober interface
func NewTCPProbe(server, port string) (Prober, error) {
	p, err := tcp.New(server, port)
	if err != nil {
		return nil, err
	}
	return p, nil
}
