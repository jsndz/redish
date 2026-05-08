package rpush

import (
	"errors"
	"fmt"

	"github.com/jsndz/redish/internal/store"
)

func Execute(args []interface{}, st *store.Store) ([]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("-ERR wrong number of arguments\r\n")
	}

	key, ok := args[0].(string)
	if !ok {
		return nil, errors.New("-ERR invalid key\r\n")
	}
	values := []string{}

	for _, value := range args[1:] {
		val, ok := value.(string)
		if !ok {
			return nil, errors.New("-ERR invalid value\r\n")
		}
		values = append(values, val)
	}

	l, err := st.Rpush(key, values...)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf(":%d\r\n", l)), nil
}
