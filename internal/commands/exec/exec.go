package exec

import (
	"errors"
	"fmt"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/configs"
	"github.com/jsndz/redish/internal/store"
)

type Dispatcher func(*client.Client, []interface{}, *store.Store, *configs.Config) ([]byte, error)

func Execute(c *client.Client, args []interface{}, st *store.Store, cfg *configs.Config, dispatch Dispatcher) ([]byte, error) {
	if len(args) != 0 {
		return nil, errors.New("-ERR wrong number of arguments\r\n")
	}
	if !c.InTx {
		return nil, errors.New("-ERR EXEC without MULTI\r\n")
	}
	if c.DirtyCAS {
		c.InTx = false
		c.TxQueue = nil
		return []byte("*0\r\n"), nil
	}
	queue := c.TxQueue
	c.InTx = false
	c.TxQueue = nil

	res := []byte(fmt.Sprintf("*%d\r\n", len(queue)))
	for _, qCmd := range queue {
		qArr := append([]interface{}{qCmd.Name}, qCmd.Args...)
		qRes, err := dispatch(c, qArr, st, cfg)
		if err != nil {
			res = append(res, []byte(err.Error())...)
		} else {
			res = append(res, qRes...)
		}
	}
	return res, nil
}
