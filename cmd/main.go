package main

import (
	"fmt"
	"net"
	"os"

	"github.com/jsndz/redish/internal/commands"
	"github.com/jsndz/redish/internal/store"
	"github.com/jsndz/redish/util"
)

func handler(conn net.Conn, st *store.Store) {
	buf := make([]byte, 1024)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		val, _ := util.RESPFormatter(string(buf[:n]))

		arr, ok := val.([]interface{})
		if !ok || len(arr) == 0 {
			conn.Write([]byte("-ERR invalid request\r\n"))
			continue
		}

		if err := commands.Dispatch(conn, arr, st); err != nil {
			conn.Write([]byte(err.Error()))
		}
	}
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379", err.Error())
		os.Exit(1)
	}
	defer l.Close()
	st := store.New()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection on 6379")
			os.Exit(1)
		}
		go handler(conn, st)
	}
}
