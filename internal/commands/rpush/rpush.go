package rpush

import (
	"errors"
	"fmt"
	"net"

	"github.com/jsndz/redish/internal/store"
)

func Execute(conn net.Conn, args []interface{}, st *store.Store) error {
	if len(args) < 2 {
		return errors.New("-ERR wrong number of arguments\r\n")
	}

	key, ok := args[0].(string)
	if !ok {
		return errors.New("-ERR invalid key\r\n")
	}
	values := []string{}

	for _, value := range args[1:] {
		val, ok := value.(string)
		if !ok {
			return errors.New("-ERR invalid value\r\n")
		}
		values = append(values, val)
	}

	l, err := st.Rpush(key, values...)
	if err != nil {
		return err
	}
	_, err = conn.Write([]byte(fmt.Sprintf(":%d\r\n", l)))

	return err
}
