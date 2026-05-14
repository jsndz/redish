package main

import (
	"github.com/jsndz/redish/internal/server"
)

func main() {
	s := server.NewServer()
	s.Start()
}
