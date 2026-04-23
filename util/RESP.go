package util

import (
	"strconv"
	"strings"
)

func RESPFormatter(msg string) (interface{}, int) {
	if len(msg) == 0 {
		return nil, 0
	}
	switch msg[0] {
	case '+':
		end := strings.Index(msg, "\r\n")
		return msg[1:end], end + 2
	case '-':
		end := strings.Index(msg, "\r\n")
		return msg[1:end], end + 2
	case ':':
		end := strings.Index(msg, "\r\n")
		n, _ := strconv.Atoi(msg[1:end])
		return n, end + 2
	case '$':
		end := strings.Index(msg, "\r\n")
		length, _ := strconv.Atoi(msg[1:end])
		if length == -1 {
			return nil, end + 2
		}
		start := end + 2
		return msg[start : start+length], start + length + 2
	case '*':
		end := strings.Index(msg, "\r\n")
		count, _ := strconv.Atoi(msg[1:end])
		idx := end + 2
		var arr []interface{}
		for i := 0; i < count; i++ {
			val, next := RESPFormatter(msg[idx:])
			arr = append(arr, val)
			idx += next
		}
		return arr, idx
	}
	return nil, 0
}
