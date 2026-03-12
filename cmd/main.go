package main

import (
	"github.com/Atennop1/atedis/internal/server"
)

func main() {
	server, err := server.New(6379)
	if err != nil {
		panic(err)
	}

	if err := server.Serve(); err != nil {
		panic(err)
	}
}
