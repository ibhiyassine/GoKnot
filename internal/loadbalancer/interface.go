package loadbalancer

import (
	"net/url"

	"github.com/ibhiyassine/GoKnot/internal/domain"
)

type LoadBalancer interface {
	GetNextValidPeer() (*domain.Backend, error)
	AddBackend(backend *domain.Backend)
	SetBackendStatus(uri *url.URL, alive bool)
	GetBackends() []*domain.Backend
	RemoveBackend(uri *url.URL)
}
