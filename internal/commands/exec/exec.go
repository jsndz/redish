package exec

import (
	"errors"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/store"
)

func Execute(c *client.Client, args []interface{}, st *store.Store) error {
	if len(args) != 0 {
		return errors.New("-ERR invalid number of args")
	}
	if !c.InTx {
		return errors.New("-ERR EXEC without MULTI")
	}
	if len(c.TxQueue) == 0 {
		c.Conn.Write([]byte("*0\r\n"))
	}
	return nil
}
