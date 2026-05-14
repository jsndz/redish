package pubsub

import (
	"errors"
	"fmt"

	"github.com/jsndz/redish/internal/channel"
	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/core"
	"github.com/jsndz/redish/internal/store"
)

func ExecutePublish(c *client.Client, args []interface{}, st *store.Store, repl *core.Replication, channels map[string]*channel.Channel,
) ([]byte, error) {

	chanName, ok := args[0].(string)
	if !ok {
		return nil, errors.New("-ERR invalid type of channel name\r\n")
	}
	message, ok := args[1].(string)
	if !ok {
		return nil, errors.New("-ERR invalid type of message\r\n")
	}
	ch, ok := channels[chanName]
	if !ok {
		ch = channel.NewChannel(chanName)
		channels[chanName] = ch
	}
	for _, subscriber := range ch.Subscribers {
		subscriber.Conn.Write([]byte(fmt.Sprintf("*3\r\n$7\r\nmessage\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(chanName), chanName, len(message), message)))
		// the output is handled by client library or cli
	}

	return []byte(fmt.Sprintf(":%d\r\n", len(ch.Subscribers))), nil
}
