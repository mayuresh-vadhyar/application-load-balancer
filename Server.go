package main

import (
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

type ServerPayload struct {
	Url    string `json:"url"`
	Weight int    `json:"weight"`
}

func CreateServer(rawUrl string) *Server {
	parsedUrl, _ := url.Parse(rawUrl)
	server := &Server{
		URL:       parsedUrl,
		IsHealthy: true,
	}
	return server
}

func CreateWeightedServer(rawUrl string, weight int) *Server {
	parsedUrl, _ := url.Parse(rawUrl)
	server := &Server{
		URL:           parsedUrl,
		Weight:        weight,
		CurrentWeight: 0,
		IsHealthy:     true,
	}
	return server
}

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
