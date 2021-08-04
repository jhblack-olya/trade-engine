package pushing

import (
	"sync"

	"github.com/siddontang/go/websocket"
)

type Client struct {
	id         int64
	conn       *websocket.Conn
	writeCh    chan interface{}
	l2ChangeCh chan *Level2Change
	sub        *subscription
	channels   map[string]struct{}
	mu         sync.Mutex
}
