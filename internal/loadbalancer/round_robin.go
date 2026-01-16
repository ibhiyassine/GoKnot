package loadbalancer

import (
	"errors"
	"sync/atomic"

	"github.com/ibhiyassine/GoKnot/internal/domain"
)

type RoundRobin struct {
	*ServerPool
	Current uint64 `json:"current"`
}

// Constructor
func NewRoundRobin(pool *ServerPool) *RoundRobin {
	return &RoundRobin{
		ServerPool: pool,
	}
}

func (r *RoundRobin) GetNextValidPeer() (*domain.Backend, error) {
	r.mux.RLock()
	defer r.mux.RUnlock()

	n := len(r.Backends)
	if n == 0 {
		return nil, errors.New("Pool doesn't contain any backend servers")
	}

	// Search for an alive server and pick it
	for range r.Backends {
		next := atomic.AddUint64(&r.Current, 1)
		idx := next % uint64(n)
		if r.Backends[idx].IsAlive() {
			return r.Backends[idx], nil
		}
	}

	return nil, errors.New("All servers in pool aren't alive")

}
