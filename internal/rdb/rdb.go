package rdb

import "encoding/hex"

// no rdb implementation yet, just a placeholder for the rdb data
// the hex string is a dummy rdb empty file
// this is a mock implementation, in a real implementation, we would read the rdb file and return its contents
func ReadRDB() []byte {
	data, err := hex.DecodeString("524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2")
	if err != nil {
		panic(err)
	}
	return data
}

func LoadRDB(data []byte) error {
	// this is a mock implementation, in a real implementation, we would read the rdb file and load its contents into the store
	_ = data
	// parse the data and load it into the store
	// something like store.Load(data)
	return nil
}
