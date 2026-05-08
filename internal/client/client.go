package client

import "net"

type Command struct {
	Name string
	Args []interface{}
}
type Client struct {
	Conn    net.Conn
	InTx    bool
	TxQueue []Command
}

func New(conn net.Conn) *Client {
	return &Client{
		Conn:    conn,
		InTx:    false,
		TxQueue: make([]Command, 0),
	}
}

func (c *Client) Write(data []byte) (int, error) {
	return c.Conn.Write(data)
}
