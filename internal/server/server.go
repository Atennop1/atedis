package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/sync/errgroup"
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
	fmt.Printf("listening on port :%d...\n", s.port)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("server: failed to create listener on port %d: %w", s.port, err)
	}

	// TODO: add context propagation
	var eg errgroup.Group

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		<-ctx.Done()
		l.Close() //nolint:errcheck
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}

			return fmt.Errorf("server: failed to establish connection on port %d, %w", s.port, err)
		}

		c := conn
		eg.Go(func() error {
			defer c.Close() //nolint:errcheck

			fmt.Println("accepted connection")
			if err := serveConnection(c); err != nil {
				return fmt.Errorf("server: failed to serve connection on port %d: %w", s.port, err)
			}

			return nil
		})
	}

	if err = eg.Wait(); err != nil {
		return fmt.Errorf("server: one of connections failed: %w", err)
	}

	fmt.Println("server gracefully stopped")
	return nil
}

func serveConnection(conn net.Conn) error {
	reader := NewRespReader(conn)
	writer := NewRespWriter(conn)

	for {
		value, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("server: connection is closed")
				return nil
			}

			return fmt.Errorf("server: failed to read from RespReader: %w", err)
		}

		if value.typ != RespTypeArray {
			fmt.Println("server: invalid request, expected array")
			continue
		}

		if len(value.array) == 0 {
			fmt.Println("server: invalid request, expected array length to be at least 1")
			continue
		}

		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		handler, ok := Handlers[command]
		if !ok {
			fmt.Printf("invalid command: %s\n", command)

			err = writer.Write(RespValue{typ: RespTypeError, str: fmt.Sprintf("unknown command: \n'%s\n'", command)})
			if err != nil {
				return fmt.Errorf("server: failed to write to RespWriter: %w", err)
			}

			continue
		}

		err = writer.Write(handler(args))
		if err != nil {
			return fmt.Errorf("server: failed to write to RespWriter: %w", err)
		}
	}
}
