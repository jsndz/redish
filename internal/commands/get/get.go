package get

import (
	"errors"
	"fmt"

	"github.com/jsndz/redish/internal/store"
)

func Execute(args []interface{}, st *store.Store) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("-ERR wrong number of arguments\r\n")
	}

	key, ok := args[0].(string)
	if !ok {
		return nil, errors.New("-ERR invalid key\r\n")
	}

	value, ok := st.Get(key)
	if !ok {
		return []byte("$-1\r\n"), nil
	}

	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)), nil
}
