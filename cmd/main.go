package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/jsndz/redish/internal/aof"
	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/commands"
	"github.com/jsndz/redish/internal/config"
	"github.com/jsndz/redish/internal/server"
	"github.com/jsndz/redish/internal/store"
	"github.com/jsndz/redish/util"
)

func handler(c *client.Client, st *store.Store, cfg *config.Config, aofHandler *aof.AOF, replication *server.Replication) {
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

		resp, err := commands.Dispatch(c, arr, st, cfg, replication)

		if err != nil {
			c.Write([]byte(err.Error()))
			continue
		}
		if replication.Role == "master" && replication.Replicas != nil && aof.IsWriteCommand(strings.ToUpper(cmdName)) {
			replication.WriteToReplicas(buf[:n])
		}
		if aofHandler != nil && aof.IsWriteCommand(strings.ToUpper(cmdName)) {
			aofHandler.Write(raw)
		}

		if !c.IsMasterConnection {
			c.Write(resp)
		}
	}
}

func main() {
	server := server.NewServer()
	if server.Master != nil {
		go handler(
			server.Master,
			server.Store,
			server.Config,
			server.Aof,
			server.Replication,
		)
	}
	c := &client.Client{}

	if server.Aof != nil {
		appendCommands := server.Aof.Restore()
		for _, cmd := range appendCommands {
			_, err := commands.Dispatch(c, cmd, server.Store, server.Config, server.Replication)
			if err != nil {
				panic(err)
			}
		}
	}
	connectionUrl := fmt.Sprintf("0.0.0.0:%d", server.Config.Port)
	l, err := net.Listen("tcp", connectionUrl)
	if err != nil {
		fmt.Println("Failed to bind to port ", err.Error())
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

		go handler(c, server.Store, server.Config, server.Aof, server.Replication)
	}
}
