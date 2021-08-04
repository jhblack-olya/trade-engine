package pushing

import (
	"sync"
)

type subscription struct {
	subscribers map[string]map[int64]*Client
	mu          sync.RWMutex
}

func newSubscription() *subscription {
	return &subscription{subscribers: map[string]map[int64]*Client{}}
}

func (s *subscription) publish(channel string, msg interface{}) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, found := s.subscribers[channel]
	if !found {
		return
	}

	for _, c := range s.subscribers[channel] {
		c.writeCh <- msg
	}
}
