package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
)

var Servers []*Server
var lb LoadBalancingStrategy

func createServer(w http.ResponseWriter, r *http.Request) {
	var newServer ServerPayload
	if decodeErr := json.NewDecoder(r.Body).Decode(&newServer); decodeErr != nil {
		WriteErrorResponse(w, http.StatusBadRequest, decodeErr.Error())
		return
	}

	for _, server := range Servers {
		if server.URL.String() == newServer.Url {
			WriteErrorResponse(w, http.StatusFound, "Server already registered")
			return
		}
	}

	server, createErr := CreateServer(newServer.Url)
	if createErr != nil {
		WriteErrorResponse(w, http.StatusBadRequest, createErr.Error())
		return
	}

	Servers = append(Servers, server)
	WriteSuccessResponse(w, http.StatusCreated, server)
}

func deleteServer(w http.ResponseWriter, r *http.Request) {
	var target ServerPayload
	if decodeErr := json.NewDecoder(r.Body).Decode(&target); decodeErr != nil {
		WriteErrorResponse(w, http.StatusBadRequest, decodeErr.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	for i, server := range Servers {
		if server.URL.String() == target.Url {
			server.stopHealthCheck()
			Servers = slices.Delete(Servers, i, i+1)
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
