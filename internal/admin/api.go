package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"github.com/ibhiyassine/GoKnot/internal/domain"
	"github.com/ibhiyassine/GoKnot/internal/loadbalancer"
)

type AdminServer struct {
	loadBalancer loadbalancer.LoadBalancer
}

func NewAdminServer(lb loadbalancer.LoadBalancer) *AdminServer {
	return &AdminServer{
		loadBalancer: lb,
	}
}

func (a *AdminServer) Start() {
	// GET /status
	http.HandleFunc("/status", a.getStatus)

	// DELETE | POST /backends
	http.HandleFunc("/backends", a.handleBackends)

	http.ListenAndServe(":8081", nil)
}

func (a *AdminServer) getStatus(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		backends := a.loadBalancer.GetBackends()
		response := map[string]any{
			"total_backends": len(backends),
			"backends":       backends,
		}

		w.Header().Set("Content-type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, "Can't retrieve backends status", http.StatusBadGateway)
			return
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	return

}

func (a *AdminServer) handleBackends(w http.ResponseWriter, r *http.Request) {
	// The body of the request will be as follow
	// {"url" : "<url_of_the_backend>"}
	var body struct {
		URL string `json:"url"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	parsedURL, err := url.Parse(body.URL)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPost:
		a.handleBackendsPost(w, parsedURL)

	case http.MethodDelete:
		a.handleBackendsDelete(w, parsedURL)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

func (a *AdminServer) handleBackendsPost(w http.ResponseWriter, uri *url.URL) {
	b := &domain.Backend{
		URL:   uri,
		Alive: true, // Default to true, HealthCheck will correct it if false
	}
	a.loadBalancer.AddBackend(b)
	log.Printf("[Admin] Added backend: %s", uri)
	w.WriteHeader(http.StatusCreated)
}

func (a *AdminServer) handleBackendsDelete(w http.ResponseWriter, uri *url.URL) {
	a.loadBalancer.RemoveBackend(uri)
	log.Printf("[Admin] Removed backend: %s", uri)
	w.WriteHeader(http.StatusOK)
}
