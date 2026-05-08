package multi

import (
	"errors"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/store"
)

func Execute(c *client.Client, args []interface{}, st *store.Store) error {
	if len(args) != 0 {
		return errors.New("-ERR no args are supported")
	}
	c.InTx = true
	_, err := c.Write([]byte("+OK\r\n"))
	return err
}
