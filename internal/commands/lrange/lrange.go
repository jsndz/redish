package lrange

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	"github.com/jsndz/redish/internal/store"
)

func Execute(args []interface{}, st *store.Store) ([]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("-ERR invalid number of args")
	}
	if len(args) != 3 {
		return nil, errors.New("-ERR invalid number of args")

	}
	//LRANGE list_key 0 1

	key, ok := args[0].(string)
	if !ok {
		return nil, errors.New("-ERR invalid key\r\n")
	}

	start, ok := args[1].(string)
	if !ok {
		return nil, errors.New("-ERR invalid key\r\n")
	}

	end, ok := args[2].(string)
	if !ok {
		return nil, errors.New("-ERR invalid key\r\n")
	}
	startInt, err := strconv.Atoi(start)
	if err != nil {
		return nil, errors.New("-ERR invalid start index\r\n")
	}
	endInt, err := strconv.Atoi(end)
	if err != nil {
		return nil, errors.New("-ERR invalid end index\r\n")
	}
	values, err := st.Lrange(key, startInt, endInt)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("*%d\r\n", len(values)))
	for _, v := range values {
		buf.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
	}
	return buf.Bytes(), nil
}
