package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/jsndz/redish/internal/aof"
	"github.com/jsndz/redish/internal/channel"
	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/commands"
	"github.com/jsndz/redish/internal/config"
	"github.com/jsndz/redish/internal/core"
	"github.com/jsndz/redish/internal/rdb"
	"github.com/jsndz/redish/internal/store"
	"github.com/jsndz/redish/util"
)

type Server struct {
	Clients     map[string]*client.Client
	Replication *core.Replication
	Store       *store.Store
	Config      *config.Config
	Master      *client.Client
	Aof         *aof.AOF
	Channels    map[string]*channel.Channel
}

func (s *Server) HandleConnection(c *client.Client) {
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

		if c.SubscribeMode {
			switch cmdName {
			case "SUBSCRIBE",
				"PSUBSCRIBE",
				"UNSUBSCRIBE",
				"PUNSUBSCRIBE",
				"PING",
				"QUIT":

			default:
				c.Write([]byte("-ERR only (P)SUBSCRIBE / (P)UNSUBSCRIBE / PING / QUIT allowed in this context\r\n"))
				continue
			}
		}

		resp, err := commands.Dispatch(c, arr, s.Store, s.Config, s.Replication, s.Channels)

		if err != nil {
			c.Write([]byte(err.Error()))
			continue
		}
		if s.Replication.Role == "master" && s.Replication.Replicas != nil && aof.IsWriteCommand(strings.ToUpper(cmdName)) {
			s.Replication.WriteToReplicas(buf[:n])

		}
		if s.Aof != nil && aof.IsWriteCommand(strings.ToUpper(cmdName)) {
			s.Aof.Write(raw)
		}

		if !c.IsMasterConnection {
			s.Replication.ReplOffset += int64(n)
			c.Write(resp)
		}
	}
}

func (s *Server) Start() {
	if s.Master != nil {
		go s.HandleConnection(s.Master)
	}
	c := &client.Client{}

	if s.Aof != nil {
		appendCommands := s.Aof.Restore()
		for _, cmd := range appendCommands {
			_, err := commands.Dispatch(c, cmd, s.Store, s.Config, s.Replication, s.Channels)
			if err != nil {
				panic(err)
			}
		}
	}
	connectionUrl := fmt.Sprintf("0.0.0.0:%d", s.Config.Port)
	l, err := net.Listen("tcp", connectionUrl)
	if err != nil {
		fmt.Println("Failed to bind to port ", err.Error())
		return
	}

	defer l.Close()

	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Failed to accept connection")
			continue
		}
		c := client.New(conn)

		go s.HandleConnection(c)
	}
}

func NewServer() *Server {

	cfg := config.SetConfig()
	aofhandler, err := aof.New(cfg)
	if err != nil {
		panic(err)
	}
	var role string
	var replicas map[string]*client.Client
	var replId string
	var replOffset int64
	var master *client.Client
	if cfg.Replicaof == "" {
		role = "master"
		replicas = make(map[string]*client.Client)
		replId, _ = util.RandString(10)
		replOffset = 0
		master = nil
	} else {
		role = "replica"
		replicas = nil
		replId = ""
		replOffset = 0
	}
	if role == "replica" {
		conn, err := net.Dial("tcp", cfg.Replicaof)
		if err != nil {
			panic(err)
		}
		master = client.New(conn)
		master.IsMasterConnection = true
		conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))

		resp := make([]byte, 1024)
		n, err := conn.Read(resp)
		if err != nil {
			panic(err)
		}

		if string(resp[:n]) != "+PONG\r\n" {
			panic("invalid response from master")
		}
		conn.Write([]byte(fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$8\r\nlistening-port\r\n$4\r\n%d\r\n", cfg.Port)))
		n, err = conn.Read(resp)
		if err != nil {
			panic(err)
		}

		if string(resp[:n]) != "+PONG\r\n" {
			panic("invalid response from master")
		}
		conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"))
		n, err = conn.Read(resp)
		if err != nil {
			panic(err)
		}

		if string(resp[:n]) != "+PONG\r\n" {
			panic("invalid response from master")
		}

		conn.Write([]byte("*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"))
		n, err = conn.Read(resp)
		if err != nil {
			panic(err)
		}
		reader := bufio.NewReader(conn)
		line, err := reader.ReadString('\n')
		var cmd string
		var offset int64
		fmt.Sscanf(line, "%s %s %d\r\n", cmd, replId, replOffset)

		if cmd != "+FULLRESYNC" {
			panic("invalid resp")
		}
		replOffset = offset

		bulkHeader, err := reader.ReadString('\n')

		if err != nil {
			panic(err)
		}
		var rdbLen int
		fmt.Sscanf(bulkHeader, "$%d\r\n", &rdbLen)
		data := make([]byte, rdbLen)
		_, err = io.ReadFull(reader, data)
		if err != nil {
			panic(err)
		}
		err = rdb.LoadRDB(data)
		if err != nil {
			panic(err)
		}

	}
	replication := &core.Replication{
		Replicas:   replicas,
		Role:       role,
		ReplID:     replId,
		ReplOffset: replOffset,
	}
	return &Server{
		Clients:     make(map[string]*client.Client),
		Store:       store.New(),
		Master:      master,
		Aof:         aofhandler,
		Config:      cfg,
		Replication: replication,
		Channels:    make(map[string]*channel.Channel),
	}
}
