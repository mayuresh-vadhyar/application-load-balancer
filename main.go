package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mayuresh-vadhyar/application-load-balancer/Response"
	"github.com/mayuresh-vadhyar/application-load-balancer/server"
)

type Server = server.Server

var lb LoadBalancingStrategy

func getServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := Response.ServerResponse{
		Status: "success",
		Data:   server.Servers,
	}
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		Response.WriteErrorResponse(w, http.StatusInternalServerError, encodeErr.Error())
	}
}

func createServer(w http.ResponseWriter, r *http.Request) {
	var newServer server.ServerPayload
	if decodeErr := json.NewDecoder(r.Body).Decode(&newServer); decodeErr != nil {
		Response.WriteErrorResponse(w, http.StatusBadRequest, decodeErr.Error())
		return
	}

	for _, item := range server.Servers {
		if item.URL.String() == newServer.Url {
			Response.WriteErrorResponse(w, http.StatusFound, "Server already registered")
			return
		}
	}

	item, createErr := server.CreateServer(newServer.Url)
	if createErr != nil {
		Response.WriteErrorResponse(w, http.StatusBadRequest, createErr.Error())
		return
	}

	server.Servers = append(server.Servers, item)
	Response.WriteSuccessResponse(w, http.StatusCreated, item)
}

func deleteServer(w http.ResponseWriter, r *http.Request) {
	var target server.ServerPayload
	if decodeErr := json.NewDecoder(r.Body).Decode(&target); decodeErr != nil {
		Response.WriteErrorResponse(w, http.StatusBadRequest, decodeErr.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	found := server.DeleteServer(target.Url)
	if found {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func serverHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getServer(w, r)
	case http.MethodPost:
		createServer(w, r)
	case http.MethodDelete:
		deleteServer(w, r)
	default:
		Response.WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method Not Allowed")
	}
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	server := lb.GetNextServer(server.Servers, r)
	if server == nil {
		Response.WriteErrorResponse(w, http.StatusServiceUnavailable, "No healthy server available")
		return
	}

	w.Header().Add("X-Forwarded-Server", server.URL.String())
	server.ReverseProxy().ServeHTTP(w, r)
}

func main() {
	config := GetConfig()
	lb = GetLoadBalancingStrategy(config.Algorithm)
	server.InitializeHealthCheckInterval(config.HealthCheckInterval)
	server.Servers = lb.CreateServerList(config)

	http.HandleFunc("/server", serverHandler)
	http.HandleFunc("/", proxyHandler)

	log.Println("Starting load balancer on port", config.Port)
	err := http.ListenAndServe(config.Port, nil)
	if err != nil {
		log.Fatalf("Error starting load balancer: %s\n", err.Error())
	}
}
