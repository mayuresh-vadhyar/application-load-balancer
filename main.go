package main

import (
	"encoding/json"
	"log"
	"net/http"
)

var Servers []*Server
var lb LoadBalancingStrategy

func createServer(w http.ResponseWriter, r *http.Request) {
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

	server, createErr := CreateServer(newServer.Url)
	if createErr != nil {
		http.Error(w, createErr.Error(), http.StatusBadRequest)
		return
	}

	Servers = append(Servers, server)

	w.WriteHeader(http.StatusCreated)
	encodeErr := json.NewEncoder(w).Encode(server)
	if encodeErr != nil {
		http.Error(w, encodeErr.Error(), http.StatusInternalServerError)
	}
}

func deleteServer(w http.ResponseWriter, r *http.Request) {
	var target ServerPayload
	decodeErr := json.NewDecoder(r.Body).Decode(&target)
	w.Header().Set("Content-Type", "application/json")
	if decodeErr != nil {
		http.Error(w, decodeErr.Error(), http.StatusBadRequest)
		return
	}

	for i, server := range Servers {
		if server.URL.String() == target.Url {
			server.stopHealthCheck()
			Servers = append(Servers[:i], Servers[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)

}

func serverHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createServer(w, r)
	case http.MethodDelete:
		deleteServer(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
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
	InitializeHealthCheckInterval(config.HealthCheckInterval)
	Servers = lb.CreateServerList(config)

	http.HandleFunc("/server", serverHandler)
	http.HandleFunc("/", proxyHandler)

	log.Println("Starting load balancer on port", config.Port)
	err := http.ListenAndServe(config.Port, nil)
	if err != nil {
		log.Fatalf("Error starting load balancer: %s\n", err.Error())
	}
}
