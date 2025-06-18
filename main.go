package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
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

	server := CreateServer(newServer.Url)

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
	Servers = lb.CreateServerList(config)

	countOfServers := len(Servers)
	interval := getHealthCheckInterval(config.HealthCheckInterval)

	for i := 0; i < countOfServers; i++ {
		go HealthCheck(Servers[i], interval)
	}

	router := mux.NewRouter()
	router.HandleFunc("/server", createServer).Methods("POST")
	router.HandleFunc("/server", deleteServer).Methods("DELETE")
	http.HandleFunc("/", proxyHandler)

	log.Println("Starting load balancer on port", config.Port)
	err := http.ListenAndServe(config.Port, nil)
	if err != nil {
		log.Fatalf("Error starting load balancer: %s\n", err.Error())
	}
}
