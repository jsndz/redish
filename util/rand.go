package util

import (
	"crypto/rand"
	"encoding/hex"
)

func RandString(n int) (string, error) {
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:n], nil
}
