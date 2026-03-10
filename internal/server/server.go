package server

import (
	"fmt"
	"net"
)

type Server struct {
	port int
}

func New(port int) *Server {
	return &Server{
		port: port,
	}
}

func (s *Server) Serve() error {
	fmt.Println("listening on port :6379...")

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("server: failed to create listener on port %d: %w", s.port, err)
	}

	conn, err := l.Accept()
	if err != nil {
		return fmt.Errorf("server: failed to establish connection on port %d, %w", s.port, err)
	}

	defer conn.Close() //nolint:errcheck

	for {
		respReader := NewRespReader(conn)

		value, err := respReader.Read()
		if err != nil {
			return fmt.Errorf("server: failed to read from resp reader on port %d: %w", s.port, err)
		}

		fmt.Println(value)
		_, _ = conn.Write([]byte("+OK\r\n"))
	}

	return nil
}
