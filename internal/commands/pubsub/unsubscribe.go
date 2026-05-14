package pubsub

import (
	"errors"
	"fmt"

	"github.com/jsndz/redish/internal/channel"
	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/core"
	"github.com/jsndz/redish/internal/store"
)

func ExecuteUnsubscribe(c *client.Client, args []interface{}, st *store.Store, repl *core.Replication, channels map[string]*channel.Channel,
) ([]byte, error) {

	for _, arg := range args {
		chanName, ok := arg.(string)
		if !ok {
			return nil, errors.New("-ERR invalid type of channel name\r\n")
		}
		ch, ok := channels[chanName]
		if !ok {
			c.Conn.Write([]byte(fmt.Sprintf("*3\r\n$11\r\nunsubscribe\r\n$%d\r\n%s\r\n:%d\r\n", len(chanName), chanName, len(c.Subscriptions))))

		}
		delete(ch.Subscribers, c.Conn.RemoteAddr().String())
		c.Subscriptions[chanName] = false
		if c.SubscribeMode && len(c.Subscriptions) == 0 {
			c.SubscribeMode = false
		}
		c.Conn.Write([]byte(fmt.Sprintf("*3\r\n$11\r\nunsubscribe\r\n$%d\r\n%s\r\n:%d\r\n", len(chanName), chanName, len(c.Subscriptions))))
	}
	return nil, nil
}
