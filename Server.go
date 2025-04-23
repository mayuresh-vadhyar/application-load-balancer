package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type Server struct {
	URL       *url.URL
	IsHealthy bool
	Mutex     sync.Mutex
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

func main() {
	config := GetConfig()
	log.Println("Starting load balancer on port: ", config.Port)
	err := http.ListenAndServe(config.Port, nil)
	if err != nil {
		log.Fatalf("Error starting load balancer: %s\n", err.Error())
	}

	servers := CreateServerList(config.Servers)
	countOfServers := len(servers)
	for i := 0; i < countOfServers; i++ {
		HealthCheck(servers[i], config.HealthCheckInterval)
	}
}
