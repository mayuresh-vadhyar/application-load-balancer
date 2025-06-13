package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type Server struct {
	URL           *url.URL
	IsHealthy     bool
	Weight        int
	CurrentWeight int
	Mutex         sync.Mutex
}

type ServerPayload struct {
	Url    string `json:"url"`
	Weight int    `json:"weight"`
}

var Servers []*Server
var lb LoadBalancingStrategy

func (s *Server) ReverseProxy() *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(s.URL)
}

func getHealthCheckInterval(healthCheckInterval string) time.Duration {
	interval, err := time.ParseDuration(healthCheckInterval)
	if err != nil {
		interval = (time.Second * 2)
	}
	return interval
}

func CreateServer(w http.ResponseWriter, r *http.Request) {
	var newServer ServerPayload
	decodeErr := json.NewDecoder(r.Body).Decode(&newServer)
	w.Header().Set("Content-Type", "application/json")
	if decodeErr != nil {
		http.Error(w, decodeErr.Error(), http.StatusBadRequest)
		return
	}

	for _, server := range Servers {
		if server.URL.String() == newServer.Url {
			http.Error(w, "Server already registered", http.StatusFound)
			return
		}
	}

	parsedUrl, _ := url.Parse(newServer.Url)
	server := &Server{
		URL:       parsedUrl,
		IsHealthy: true,
	}

	Servers = append(Servers, server)
	w.WriteHeader(http.StatusCreated)
	encodeErr := json.NewEncoder(w).Encode(server)
	if encodeErr != nil {
		http.Error(w, encodeErr.Error(), http.StatusInternalServerError)
	}
}

func DeleteServer(w http.ResponseWriter, r *http.Request) {
	var target ServerPayload
	decodeErr := json.NewDecoder(r.Body).Decode(&target)
	w.Header().Set("Content-Type", "application/json")
	if decodeErr != nil {
		http.Error(w, decodeErr.Error(), http.StatusBadRequest)
		return
	}

	serverFound := false
	for i, server := range Servers {
		if server.URL.String() == target.Url {
			serverFound = true
			Servers = append(Servers[:i], Servers[i+1:]...)
			break
		}
	}

	if serverFound {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	server := lb.GetNextServer(Servers, r)
	if server == nil {
		http.Error(w, "No healthy server available", http.StatusServiceUnavailable)
		return
	}

	w.Header().Add("X-Forwarded-Server", server.URL.String())
	server.ReverseProxy().ServeHTTP(w, r)
}

func main() {
	config := GetConfig()
	lb = GetLoadBalancingStrategy(config.Algorithm)
	Servers = lb.CreateServerList(config)

	countOfServers := len(Servers)
	interval := getHealthCheckInterval(config.HealthCheckInterval)

	for i := 0; i < countOfServers; i++ {
		go HealthCheck(Servers[i], interval)
	}

	router := mux.NewRouter()
	router.HandleFunc("/server", CreateServer).Methods("POST")
	router.HandleFunc("/server", DeleteServer).Methods("DELETE")
	router.HandleFunc("/", ProxyHandler)

	log.Println("Starting load balancer on port", config.Port)
	err := http.ListenAndServe(config.Port, nil)
	if err != nil {
		log.Fatalf("Error starting load balancer: %s\n", err.Error())
	}
}
