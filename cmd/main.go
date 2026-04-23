package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/jsndz/redish/util"
)

func handler(conn net.Conn, MAP map[string]string) {
	buf := make([]byte, 1024)
	var mu sync.Mutex

	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("connection closed")
			return
		}
		val, _ := util.RESPFormatter(string(buf[:n]))

		arr, ok := val.([]interface{})
		if !ok || len(arr) == 0 {
			conn.Write([]byte("-ERR invalid request\r\n"))
			return
		}

		cmd, ok := arr[0].(string)
		if !ok {
			conn.Write([]byte("-ERR invalid command\r\n"))
			return
		}

		switch strings.ToUpper(cmd) {

		case "PING":
			conn.Write([]byte("+PONG\r\n"))

		case "ECHO":
			if len(arr) < 2 {
				conn.Write([]byte("-ERR wrong number of arguments\r\n"))
				return
			}
			msg, ok := arr[1].(string)
			if !ok {
				conn.Write([]byte("-ERR invalid argument\r\n"))
				return
			}
			resp := "$" + strconv.Itoa(len(msg)) + "\r\n" + msg + "\r\n"
			conn.Write([]byte(resp))
		case "SET":
			mu.Lock()
			MAP[arr[1].(string)] = arr[2].(string)
			mu.Unlock()
			resp := "+OK\r\n"
			conn.Write([]byte(resp))
		case "GET":
			mu.Lock()
			data, ok := MAP[arr[1].(string)]
			mu.Unlock()
			if !ok {
				conn.Write([]byte("$-1\r\n"))
				return
			}
			resp := "$" + strconv.Itoa(len(data)) + "\r\n" + data + "\r\n"

			conn.Write([]byte(resp))
		default:
			conn.Write([]byte("-ERR unknown command\r\n"))
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
