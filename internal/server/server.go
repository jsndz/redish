package server

import (
	"bufio"
	"fmt"
	"io"
	"net"

	"github.com/jsndz/redish/internal/aof"
	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/config"
	"github.com/jsndz/redish/internal/rdb"
	"github.com/jsndz/redish/internal/store"
	"github.com/jsndz/redish/util"
)

type Replication struct {
	Role       string
	ReplID     string
	ReplOffset int64
	Replicas   map[string]*client.Client
}

func (r *Replication) WriteToReplicas(raw []byte) {
	for _, replica := range r.Replicas {
		replica.Conn.Write(raw)
	}
	r.ReplOffset += int64(len(raw))
}

type Server struct {
	Clients     map[string]*client.Client
	Replication *Replication
	Store       *store.Store
	Config      *config.Config
	MasterPort  int
	MasterHost  string
	Aof         *aof.AOF
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
	if cfg.Replicaof == "" {
		role = "master"
		replicas = make(map[string]*client.Client)
		replId, _ = util.RandString(10)
		replOffset = 0
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
	replication := &Replication{
		Replicas:   replicas,
		Role:       role,
		ReplID:     replId,
		ReplOffset: replOffset,
	}
	return &Server{
		Clients:     make(map[string]*client.Client),
		Store:       store.New(),
		MasterPort:  cfg.Port,
		MasterHost:  "localhost",
		Aof:         aofhandler,
		Config:      cfg,
		Replication: replication,
	}
}
