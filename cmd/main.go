package main

import (
	"github.com/Atennop1/atedis/internal/server"
)

func main() {
	server := server.New(6379)
	if err := server.Serve(); err != nil {
		panic(err)
	}
}
