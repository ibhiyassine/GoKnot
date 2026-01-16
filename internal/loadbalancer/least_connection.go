package loadbalancer

import (
	"errors"
	"math"
	"sync/atomic"

	"github.com/ibhiyassine/GoKnot/internal/domain"
)

type LeastConnections struct {
	*ServerPool
}

func NewLeastConnections(pool *ServerPool) *LeastConnections {
	return &LeastConnections{
		ServerPool: pool,
	}
}

func (l *LeastConnections) GetNextValidPeer() (*domain.Backend, error) {
	l.mux.RLock()
	defer l.mux.RUnlock()

	var best *domain.Backend
	var min int64 = math.MaxInt64

	for _, b := range l.Backends {
		if !b.IsAlive() {
			continue
		}
		conn := atomic.LoadInt64(&b.CurrentConns)

		if conn < min {
			min = conn
			best = b
		}
	}

	if min == math.MaxInt64 {
		return nil, errors.New("All servers in pool aren't alive")
	} else {
		return best, nil

	}
}
