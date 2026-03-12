package server

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type AOF struct {
	file *os.File
	mu   sync.Mutex
}

func NewAOF(path string) (*AOF, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("aof: failed to open file %s: %w", path, err)
	}

	aof := &AOF{
		file: f,
		mu:   sync.Mutex{},
	}

	// background goroutine for syncing file
	go func() {
		aof.mu.Lock()
		aof.file.Sync() //nolint:errcheck
		aof.mu.Unlock()

		time.Sleep(1 * time.Second)
	}()

	return aof, nil
}

func (a *AOF) Write(v RespValue) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	_, err := a.file.Write(v.Marshal())
	if err != nil {
		return fmt.Errorf("aof: failed to write: %w", err)
	}

	return nil
}

func (a *AOF) Read() (RespValue, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	reader := NewRespReader(a.file)
	result := []RespValue{}

	for {
		value, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return RespValue{}, fmt.Errorf("aof: failed to read from aof: %w", err)
		}

		result = append(result, value)
	}

	return RespValue{typ: RespTypeArray, array: result}, nil
}

func (a *AOF) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return fmt.Errorf("aof: failed to close: %w", a.file.Close())
}
