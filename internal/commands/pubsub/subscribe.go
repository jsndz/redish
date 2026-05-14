package pubsub

import (
	"errors"
	"fmt"

	"github.com/jsndz/redish/internal/channel"
	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/core"
	"github.com/jsndz/redish/internal/store"
)

func ExecuteSubscribe(c *client.Client, args []interface{}, st *store.Store, repl *core.Replication, channels map[string]*channel.Channel,
) ([]byte, error) {

	for _, arg := range args {
		chanName, ok := arg.(string)
		if !ok {
			return nil, errors.New("-ERR invalid type of channel name\r\n")
		}
		ch, ok := channels[chanName]
		if !ok {
			ch = channel.NewChannel(chanName)
			channels[chanName] = ch
		}
		ch.Subscribers[c.Conn.RemoteAddr().String()] = c
		c.Subscriptions[chanName] = true
		c.SubscribeMode = true
		c.Conn.Write([]byte(fmt.Sprintf("*3\r\n$9\r\nsubscribe\r\n$%d\r\n%s\r\n:%d\r\n", len(chanName), chanName, len(c.Subscriptions))))
	}
	return nil, nil
}
