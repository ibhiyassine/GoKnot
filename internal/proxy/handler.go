package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/ibhiyassine/GoKnot/internal/loadbalancer"
)

type ProxyHandler struct {
	loadBalancer loadbalancer.LoadBalancer
}

func NewProxyHandler(lb loadbalancer.LoadBalancer) *ProxyHandler {
	return &ProxyHandler{
		loadBalancer: lb,
	}
}

func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	peer, err := ph.loadBalancer.GetNextValidPeer()

	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	targetURL := peer.URL
	log.Printf("Proxy requesting to %s", targetURL.String())

	// The peer will get a connection
	peer.IncrementConns()
	defer peer.DecrementConns()

	// setup the reverse proxy
	proxy := ph.getReverseProxy(targetURL)

	// The request context is passed
	proxy.ServeHTTP(w, r)

}

func (ph *ProxyHandler) getReverseProxy(uri *url.URL) (proxy *httputil.ReverseProxy) {
	proxy = httputil.NewSingleHostReverseProxy(uri)

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		// If this function gets triggered, that means the backend isn't suitable for requests
		log.Printf("[%s] Connection failed: %v", uri, err)

		// It should be marked as dead
		ph.loadBalancer.SetBackendStatus(uri, false)

		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	return
}
