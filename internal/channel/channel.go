package channel

import "github.com/jsndz/redish/internal/client"

type Channel struct {
	Name        string
	Subscribers map[string]*client.Client
}

func NewChannel(name string) *Channel {
	return &Channel{
		Name:        name,
		Subscribers: make(map[string]*client.Client),
	}
}
