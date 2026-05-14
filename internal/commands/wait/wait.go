package wait

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/core"
	"github.com/jsndz/redish/internal/store"
	"github.com/jsndz/redish/util"
)

func Execute(c *client.Client, args []interface{}, st *store.Store, repl *core.Replication) ([]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("-ERR invalid number of args")
	}
	replicasRequired, ok := args[0].(int)
	if !ok {
		return nil, errors.New("-ERR invalid arg")
	}
	timeout, ok := args[1].(int)
	if !ok {
		return nil, errors.New("-ERR invalid arg")
	}
	if len(repl.Replicas) == 0 || replicasRequired == 0 {
		return []byte(":0\r\n"), nil
	}

	if repl.ReplOffset == 0 {
		return []byte(fmt.Sprintf(":%d\r\n", len(repl.Replicas))), nil
	}
	targetOffset := repl.ReplOffset
	ch := make(chan int)
	for _, replica := range repl.Replicas {

		go func(rep *client.Client) {
			buf := make([]byte, 1024)
			rep.Conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$6\r\nGETACK\r\n$1\r\n*\r\n"))
			n, err := rep.Conn.Read(buf)
			if err != nil {
				return
			}
			raw := string(buf[:n])
			val, _ := util.RESPFormatter(raw)

			arr, ok := val.([]interface{})
			if !ok || len(arr) != 3 {
				return
			}

			cmd, _ := arr[0].(string)
			subcmd, _ := arr[1].(string)
			offsetStr, ok := arr[2].(string)
			if !ok {
				return
			}

			offset, err := strconv.ParseInt(offsetStr, 10, 64)
			if err != nil {
				return
			}
			if cmd == "REPLCONF" && subcmd == "ACK" && offset >= targetOffset {
				ch <- 1
			}
		}(replica)
	}

	ack := 0
	timer := time.After(time.Duration(timeout) * time.Millisecond)

	for range repl.Replicas {
		select {
		case <-ch:
			ack++
			if ack >= replicasRequired {
				return []byte(fmt.Sprintf(":%d\r\n", ack)), nil
			}
		case <-timer:
			return []byte(fmt.Sprintf(":%d\r\n", ack)), nil
		}
	}
	return []byte(fmt.Sprintf(":%d\r\n", ack)), nil
}
