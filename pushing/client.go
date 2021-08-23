package pushing

import "sync"

// Each connection corresponds to a client, and the client is responsible for the data I / O of the connection
type Client struct {
	id int64
	// conn       *websocket.Conn
	writeCh chan interface{}
	// l2ChangeCh chan *Level2Change
	sub      *subscription
	channels map[string]struct{}
	mu       sync.Mutex
}
