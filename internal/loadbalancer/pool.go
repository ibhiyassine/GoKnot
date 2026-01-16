package loadbalancer

import (
	"net/url"
	"sync"

	"github.com/ibhiyassine/GoKnot/internal/domain"
)

type ServerPool struct {
	Backends []*domain.Backend `json:"backends"`
	mux      sync.RWMutex
}

func (s *ServerPool) AddBackend(backend *domain.Backend) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.Backends = append(s.Backends, backend)
}

func (s *ServerPool) SetBackendStatus(uri *url.URL, alive bool) {
	s.mux.RLock() // we are only going to read the struct
	defer s.mux.RUnlock()
	for _, b := range s.Backends {
		if b.URL.String() == uri.String() {
			b.SetAlive(alive)
		}
	}
}
