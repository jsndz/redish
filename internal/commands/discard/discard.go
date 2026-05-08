package discard

import (
	"errors"

	"github.com/jsndz/redish/internal/client"
)

func Execute(c *client.Client, args []interface{}) ([]byte, error) {
	if len(args) != 0 {
		return nil, errors.New("-ERR wrong number of arguments\r\n")
	}
	if !c.InTx {
		return nil, errors.New("-ERR DISCARD without MULTI\r\n")
	}
	c.InTx = false
	c.TxQueue = nil
	return []byte("+OK\r\n"), nil
}
