package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Server struct {
	URL           *url.URL
	IsHealthy     bool
	Weight        int
	CurrentWeight int
	Mutex         sync.Mutex
}

func (s *Server) ReverseProxy() *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(s.URL)
}

func CreateServerList(serverUrls []string) []*Server {
	var servers []*Server

	for _, rawUrl := range serverUrls {
		parsedUrl, err := url.Parse(rawUrl)
		if err != nil {
			continue
		}

		server := &Server{
			URL:       parsedUrl,
			IsHealthy: true,
		}

		servers = append(servers, server)
	}

	return servers
}

func CreateServerListForWRR(serverUrls []string, weights []int) []*Server {
	var servers []*Server
	countOfServers := len(serverUrls)

	for i := 0; i < countOfServers; i++ {
		parsedUrl, err := url.Parse(serverUrls[i])
		log.Println(parsedUrl)
		if err != nil {
			continue
		}

		server := &Server{
			URL:           parsedUrl,
			Weight:        weights[i],
			CurrentWeight: 0,
			IsHealthy:     true,
		}

		servers = append(servers, server)
	}

	return servers
}

func getHealthCheckInterval(healthCheckInterval string) time.Duration {
	interval, err := time.ParseDuration(healthCheckInterval)
	if err != nil {
		interval = (time.Second * 2)
	}
	return interval
}

func getLoadBalancingStrategy(algorithm string) LoadBalancingStrategy {
	if algorithm == "" {
		algorithm = "RR"
	}

	switch algorithm {
	case "RR":
		return &RoundRobinStrategy{Current: -1}
	case "WRR":
		return &WeightedRoundRobinStrategy{Current: -1}
	default:
		log.Fatalf("Unknown algorithm: %s", algorithm)
	}

	return nil
}

func main() {
	config := GetConfig()
	lb := getLoadBalancingStrategy(config.Algorithm)
	servers := lb.CreateServerList(config)

	countOfServers := len(servers)
	interval := getHealthCheckInterval(config.HealthCheckInterval)

	for i := 0; i < countOfServers; i++ {
		go HealthCheck(servers[i], interval)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		server := lb.GetNextServer(servers)
		if server == nil {
			http.Error(w, "No healthy server available", http.StatusServiceUnavailable)
			return
		}

		w.Header().Add("X-Forwarded-Server", server.URL.String())
		server.ReverseProxy().ServeHTTP(w, r)
	})

	log.Println("Starting load balancer on port", config.Port)
	err := http.ListenAndServe(config.Port, nil)
	if err != nil {
		log.Fatalf("Error starting load balancer: %s\n", err.Error())
	}
}
