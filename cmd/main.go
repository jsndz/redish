package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/commands"
	"github.com/jsndz/redish/internal/commands/aof"
	"github.com/jsndz/redish/internal/configs"
	"github.com/jsndz/redish/internal/store"
	"github.com/jsndz/redish/util"
)

func handler(c *client.Client, st *store.Store, cfg *configs.Config) {
	buf := make([]byte, 4096)

	for {
		n, err := c.Conn.Read(buf)
		if err != nil {
			return
		}

		raw := string(buf[:n])

		val, _ := util.RESPFormatter(raw)

		arr, ok := val.([]interface{})
		if !ok || len(arr) == 0 {
			c.Write([]byte("-ERR invalid request\r\n"))
			continue
		}

		cmdName, ok := arr[0].(string)
		if !ok {
			c.Write([]byte("-ERR invalid command\r\n"))
			continue
		}

		resp, err := commands.Dispatch(c, arr, st, cfg)

		if err != nil {
			c.Write([]byte(err.Error()))
			continue
		}

		if cfg.AppendOnly && aof.IsWriteCommand(strings.ToUpper(cmdName)) {
			cfg.WriteToAppendFile(raw)
		}

		c.Write(resp)
	}
}
func main() {
	cfg := configs.SetConfig()
	appendCommands := cfg.RestoreFromAppendFile()
	st := store.New()
	c := &client.Client{}

	for _, cmd := range appendCommands {
		_, err := commands.Dispatch(c, cmd, st, cfg)
		if err != nil {
			panic(err)
		}
	}
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379", err.Error())
		os.Exit(1)
	}

	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection on 6379")
			os.Exit(1)
		}
		c := client.New(conn)
		go handler(c, st, cfg)
	}
}
