package client

import "net"

type Client struct {
	Conn    net.Conn
	InTx    bool
	TxQueue [][]interface{}
}

func New(conn net.Conn) *Client {
	return &Client{
		Conn:    conn,
		InTx:    false,
		TxQueue: make([][]interface{}, 0),
	}
}

func (c *Client) Write(data []byte) (int, error) {
	return c.Conn.Write(data)
}
