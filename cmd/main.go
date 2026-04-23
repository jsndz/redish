package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func handler(conn net.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("connection closed")
			return
		}
		msg := strings.TrimSpace(string(buf[:n]))
		fmt.Println("received:", msg)

		switch msg {
		case "PING":
			conn.Write([]byte("+PONG\r\n"))
		default:
			fmt.Println("Unknown Command")
		}
	}
}

func main() {
	// binds program to network address and listen
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Failed to accpet 6379")
			os.Exit(1)
		}
		go handler(conn)
	}
}
