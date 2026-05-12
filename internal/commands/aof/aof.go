package aof

var writeCommands = map[string]bool{
	"SET":     true,
	"INCR":    true,
	"RPUSH":   true,
	"MULTI":   true,
	"EXEC":    true,
	"DISCARD": true,
}

func IsWriteCommand(cmd string) bool {
	return writeCommands[cmd]
}
