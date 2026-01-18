package health

import (
	"log"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/ibhiyassine/GoKnot/internal/domain"
	"github.com/ibhiyassine/GoKnot/internal/loadbalancer"
)

const DEFAULT_TIMEOUT time.Duration = 2 * time.Second

type HealthChecker struct {
	Interval time.Duration
	Timeout  time.Duration
	LB       loadbalancer.LoadBalancer
	checking bool // to check if I am currently checking the health
	mux      sync.RWMutex
}

func NewHealthChecker(lb loadbalancer.LoadBalancer, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		Interval: interval,
		LB:       lb,
		Timeout:  DEFAULT_TIMEOUT,
	}
}

func (hc *HealthChecker) Start() {
	ticker := time.NewTicker(hc.Interval)

	// create a background process
	go func() {
		log.Printf("Health ckecking started with interval %v", hc.Interval)
		for {
			select {
			case <-ticker.C:
				hc.checkHealth()
			}
		}
	}()
}

func (hc *HealthChecker) checkHealth() {
	// If it succeded we can start checking the health
	// otherwise, it is already locked we can't check right now we just wait for the next call
	if hc.mux.TryLock() {
		defer hc.mux.Unlock()
		wg := sync.WaitGroup{}
		// Iterate through all the backends
		for _, backend := range hc.LB.GetBackends() {
			wg.Add(1)
			go func(backend *domain.Backend) {
				defer wg.Done()
				alive := hc.ping(backend.URL)
				if backend.IsAlive() != alive {
					if alive {
						log.Printf("[Health] Backend %s is UP and RUNNING", backend.URL)
					} else {
						log.Printf("[Health] Backend %s is DOWN", backend.URL)
					}
				}
				backend.SetAlive(alive)
			}(backend)
		}
		wg.Wait()
	} else {
		// Already checking
		return
	}
}

func (hc *HealthChecker) ping(uri *url.URL) bool {
	// Do a TCP dial and return if it is alive or not
	conn, err := net.DialTimeout("tcp", uri.Host, hc.Timeout)
	// the dial wasn't succesful
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
