package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"

	"github.com/mayuresh-vadhyar/application-load-balancer/Response"
	"github.com/mayuresh-vadhyar/application-load-balancer/config"
	"github.com/mayuresh-vadhyar/application-load-balancer/rateLimiter"
	"github.com/mayuresh-vadhyar/application-load-balancer/server"
)

type Server = server.Server

var lb LoadBalancingStrategy

func getServer(w http.ResponseWriter, r *http.Request) {
	filteredServers := []*Server{}
	query := r.URL.Query()
	isHealthyParam := query.Get("isHealthy")
	urlParam := query.Get("urlParam")

	for _, s := range server.Servers {
		match := true

		if isHealthyParam != "" {
			wantHealthy := isHealthyParam == "true"
			if s.IsHealthy != wantHealthy {
				match = false
			}
		}

		if urlParam != "" && s.URL.String() != urlParam {
			match = false
		}

		if match {
			filteredServers = append(filteredServers, s)
		}
	}

	Response.WriteSuccessResponseArray(w, http.StatusOK, filteredServers)
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

func generateHash() (string, error) {
	hashLength := 12
	randomBytes := make([]byte, hashLength)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	hash := base64.RawURLEncoding.EncodeToString(randomBytes)
	if len(hash) > hashLength {
		hash = hash[:hashLength]
	}
	return hash, nil
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	server := lb.GetNextServer(server.Servers, r)
	if server == nil {
		Response.WriteErrorResponse(w, http.StatusServiceUnavailable, "No healthy server available")
		return
	}

	hash, _ := generateHash()
	r.Header.Add("tracking-id", hash)
	w.Header().Add("tracking-id", hash)
	w.Header().Add("X-Forwarded-Server", server.URL.String())
	server.ReverseProxy().ServeHTTP(w, r)
}

func main() {
	config := config.GetConfig()
	lb = GetLoadBalancingStrategy(config.Algorithm)
	InitializeLogResponseWriter(config.DisableLogs)
	server.InitializeHealthCheckConfig(config.HealthCheck)
	server.StartServerPoolLogRoutine()
	server.Servers = lb.CreateServerList(config)
	rl := rateLimiter.GetRateLimiter()

	http.Handle("/", loggingMiddleware(http.HandlerFunc(proxyHandler)))
	http.HandleFunc("/server", serverHandler)

	log.Println("Starting load balancer on port", config.Port)
	var err error
	if rl == nil {
		err = http.ListenAndServe(config.Port, nil)
	} else {
		err = http.ListenAndServe(config.Port, rl.RateLimit(http.DefaultServeMux))
	}
	if err != nil {
		log.Fatalf("Error starting load balancer: %s\n", err.Error())
	}
}
