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
	aof  *AOF
}

func New(port int) (*Server, error) {
	aof, err := NewAOF("aof.txt")
	if err != nil {
		return nil, fmt.Errorf("server: failed to create AOF: %w", err)
	}

	return &Server{
		port: port,
		aof:  aof,
	}, nil
}

func (s *Server) Serve() error {
	fmt.Printf("listening on port :%d...\n", s.port)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("server: failed to create listener on port %d: %w", s.port, err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// TODO: add context propagation
	eg, ctx := errgroup.WithContext(ctx)

	go func() {
		<-ctx.Done()
		l.Close() //nolint:errcheck
	}()

	actions, err := s.aof.Read()
	if err != nil {
		return fmt.Errorf("failed to read actions from AOF: %w", err)
	}

	for _, action := range actions.array {
		_, err = s.executeCommand(action)
		if err != nil {
			return fmt.Errorf("failed to execute action from AOF: %w", err)
		}
	}

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
			if err := s.serveConnection(c); err != nil {
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

func (s *Server) serveConnection(conn net.Conn) error {
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

		result, err := s.executeCommand(value)
		if err != nil {
			return fmt.Errorf("server: failed to execute command: %w", err)
		}

		// TODO: write not all commands but only meaningful ones
		err = s.aof.Write(value)
		if err != nil {
			return fmt.Errorf("server: failed to write command to AOF: %w", err)
		}

		err = writer.Write(result)
		if err != nil {
			return fmt.Errorf("server: failed to write command: %w", err)
		}
	}
}

func (s *Server) executeCommand(v RespValue) (RespValue, error) {
	if v.typ != RespTypeArray {
		return RespValue{}, fmt.Errorf("server: invalid request, expected array")
	}

	if len(v.array) == 0 {
		return RespValue{}, fmt.Errorf("server: invalid request, expected array length to be at least 1")
	}

	command := strings.ToUpper(v.array[0].bulk)
	args := v.array[1:]

	handler, ok := Handlers[command]
	if !ok {
		fmt.Printf("invalid command: %s\n", command)

		// TODO: TEMP THING FOR TESTING
		return RespValue{typ: RespTypeString, str: "OK"}, nil

		// later replace to:
		// return RespValue{typ: RespTypeError, str: fmt.Sprintf("unknown command: \n'%s\n'", command)}
	}

	return handler(args), nil
}
