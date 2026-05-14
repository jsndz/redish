package core

import "github.com/jsndz/redish/internal/client"

type Replication struct {
	Role       string
	ReplID     string
	ReplOffset int64
	Replicas   map[string]*client.Client
}

func (r *Replication) WriteToReplicas(raw []byte) {
	for _, replica := range r.Replicas {
		replica.Conn.Write(raw)
	}
	r.ReplOffset += int64(len(raw))
}
