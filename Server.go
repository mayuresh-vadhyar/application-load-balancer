package main

import (
	"context"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Server struct {
	URL             *url.URL
	IsHealthy       bool
	Weight          int
	CurrentWeight   int
	Mutex           sync.Mutex
	stopHealthCheck context.CancelFunc
}

type ServerPayload struct {
	Url    string `json:"url"`
	Weight int    `json:"weight"`
}

var interval time.Duration

func CreateServer(rawUrl string) (*Server, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	server := &Server{
		URL:       parsedUrl,
		IsHealthy: true,
	}
	return server, nil
}

func CreateWeightedServer(rawUrl string, weight int) (*Server, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	server := &Server{
		URL:           parsedUrl,
		Weight:        weight,
		CurrentWeight: 0,
		IsHealthy:     true,
	}
	return server, nil
}

func (s *Server) ReverseProxy() *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(s.URL)
}

func getHealthCheckInterval(healthCheckInterval string) time.Duration {
	once.Do(func() {
		var err error
		interval, err = time.ParseDuration(healthCheckInterval)
		if err != nil {
			interval = (time.Second * 2)
		}
	})

	return interval
}
