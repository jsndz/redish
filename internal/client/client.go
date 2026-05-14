package client

import (
	"net"
)

type Command struct {
	Name string
	Args []interface{}
}
type Client struct {
	Conn               net.Conn
	IsMasterConnection bool
	InTx               bool
	TxQueue            []Command
	WatchedKeys        map[string]bool
	DirtyCAS           bool
	Subscriptions      map[string]bool
	SubscribeMode      bool
}

func New(conn net.Conn) *Client {
	return &Client{
		Conn:               conn,
		InTx:               false,
		TxQueue:            make([]Command, 0),
		WatchedKeys:        make(map[string]bool, 0),
		IsMasterConnection: false,
		Subscriptions:      make(map[string]bool, 0),
		SubscribeMode:      false,
	}
}

func (c *Client) Write(data []byte) (int, error) {
	return c.Conn.Write(data)
}
