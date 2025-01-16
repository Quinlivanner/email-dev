package service

import (
	"email/global"
	"net"
	"sync"

	"github.com/emersion/go-smtp"
)

type limitListener struct {
	net.Listener
	maxConnections int
	current        int
	mu             sync.Mutex
}

func SmtpListenAndServe(s *smtp.Server, maxConnections int) error {
	network := network(s)

	addr := s.Addr
	if !s.LMTP && addr == "" {
		addr = ":smtp"
	}

	listener, err := net.Listen(network, addr)
	if err != nil {
		return err
	}

	limitedListener := newLimitListener(listener, maxConnections)

	return s.Serve(limitedListener)
}

func network(s *smtp.Server) string {
	if s.Network != "" {
		return s.Network
	}
	if s.LMTP {
		return "unix"
	}
	return "tcp"
}

func newLimitListener(inner net.Listener, max int) *limitListener {
	return &limitListener{
		Listener:       inner,
		maxConnections: max,
	}
}

func (l *limitListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.current >= l.maxConnections {
		global.Log.Infof("max connections limit reached")
		conn.Close()
		return conn, nil
	}

	l.current++
	return &limitConn{Conn: conn, limitListener: l}, nil
}

func (l *limitListener) release() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.current--
}

type limitConn struct {
	net.Conn
	limitListener *limitListener
	mu            sync.Mutex
	close         bool
}

func (c *limitConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.close {
		return nil
	}
	err := c.Conn.Close()
	c.close = true
	c.limitListener.release()
	return err
}
