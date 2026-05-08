package incr

import (
	"errors"
	"fmt"

	"github.com/jsndz/redish/internal/store"
)

func Execute(args []interface{}, st *store.Store) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("-ERR invalid number of arguments")
	}

	key, ok := args[0].(string)
	if !ok {
		return nil, errors.New("-ERR invalid type of key")
	}
	val, err := st.Incr(key)
	if err != nil {
		return nil, errors.New("-ERR Could not increment value: " + err.Error())
	}
	return []byte(fmt.Sprintf(":%d\r\n", val)), nil
}
