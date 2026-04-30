package get

import (
	"errors"
	"fmt"
	"net"

	"github.com/jsndz/redish/internal/store"
)

func Execute(conn net.Conn, args []interface{}, st *store.Store) error {
	if len(args) != 1 {
		return errors.New("-ERR wrong number of arguments\r\n")
	}

	key, ok := args[0].(string)
	if !ok {
		return errors.New("-ERR invalid key\r\n")
	}

	value, ok := st.Get(key)
	if !ok {
		_, err := conn.Write([]byte("$-1\r\n"))
		return err
	}

	_, err := conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)))
	return err
}
