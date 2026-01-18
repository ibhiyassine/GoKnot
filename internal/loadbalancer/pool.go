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

func (s *ServerPool) GetBackends() []*domain.Backend {
	s.mux.RLock()
	defer s.mux.RUnlock()

	// We return a copy
	list := make([]*domain.Backend, len(s.Backends))
	copy(list, s.Backends)
	return list
}

func (s *ServerPool) RemoveBackend(uri *url.URL) {
	s.mux.Lock()
	defer s.mux.Unlock()

	for i, b := range s.Backends {
		if b.URL.String() == uri.String() {
			s.Backends = append(s.Backends[:i], s.Backends[i+1:]...)
			return
		}
	}
}
