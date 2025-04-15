package main

import (
    "net/http"
    "net/url"
    "sync"
    "log"
    "net/http/httputil"
)

type Server struct {
    URL       *url.URL
    IsHealthy bool
    Mutex     sync.Mutex
}

func (s *Server) ReverseProxy() *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(s.URL)
}

func main() {
    config := GetConfig()
    log.Println("Starting load balancer on port: ", config.Port)
    err := http.ListenAndServe(config.Port, nil)
    if err != nil {
            log.Fatalf("Error starting load balancer: %s\n", err.Error())
    }

    servers := config.Servers
    countOfServers := len(servers)
    for i:=0; i < countOfServers; i++ {
        HealthCheck(servers[i], config.HealthCheckInterval)
    }
}