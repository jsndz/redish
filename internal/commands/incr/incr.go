package incr

import (
	"errors"
	"fmt"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/store"
)

func Execute(c *client.Client, args []interface{}, st *store.Store) error {
	if len(args) != 1 {
		return errors.New("-ERR invalid number of arguments")
	}

	key, ok := args[0].(string)
	if !ok {
		return errors.New("-ERR invalid type of key")
	}
	val, err := st.Incr(key)
	if err != nil {
		return errors.New("-ERR Could not increment value: " + err.Error())
	}
	_, err = c.Write([]byte(fmt.Sprintf(":%d\r\n", val)))
	return err
}
