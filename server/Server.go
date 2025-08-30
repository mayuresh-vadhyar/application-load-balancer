package server

import (
	"context"
	"encoding/json"
	"net/http/httputil"
	"net/url"
	"slices"
	"sync"
	"time"
)

type Server struct {
	Id              int                `json:"id"`
	URL             *url.URL           `json:"-"`
	IsHealthy       bool               `json:"isHealthy"`
	Weight          int                `json:"weight,omitempty"`
	CurrentWeight   int                `json:"-"`
	Mutex           sync.Mutex         `json:"-"`
	StopHealthCheck context.CancelFunc `json:"-"`
	UnhealthyChecks int8               `json:"unhealthyCHecks"`
}

func (m Server) MarshalJSON() ([]byte, error) {
	type Alias Server
	return json.Marshal(&struct {
		Alias
		URL string `json:"url"`
	}{
		Alias: (Alias)(m),
		URL:   m.URL.Host,
	})
}

type ServerPayload struct {
	Url    string `json:"url"`
	Weight int    `json:"weight"`
}

var Servers []*Server
var interval time.Duration
var intervalOnce sync.Once
var maxUnhealthyChecks int8 = -1
var idMutex sync.Mutex
var lastId int = 0

func getNextId() int {
	idMutex.Lock()
	defer idMutex.Unlock()
	lastId++
	return lastId
}

func CreateServer(rawUrl string) (*Server, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	server := &Server{
		Id:              getNextId(),
		URL:             parsedUrl,
		IsHealthy:       true,
		StopHealthCheck: cancel,
	}
	go StartHealthCheckRoutine(ctx, server, interval)

	return server, nil
}

func DeleteServer(targetUrl string) bool {
	for i, item := range Servers {
		if item.URL.String() == targetUrl {
			item.StopHealthCheck()
			Servers = slices.Delete(Servers, i, i+1)
			return true
		}
	}
	return false
}

func CreateWeightedServer(rawUrl string, weight int) (*Server, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	server := &Server{
		Id:              getNextId(),
		URL:             parsedUrl,
		Weight:          weight,
		CurrentWeight:   0,
		IsHealthy:       true,
		StopHealthCheck: cancel,
	}
	go StartHealthCheckRoutine(ctx, server, interval)

	return server, nil
}

func (s *Server) ReverseProxy() *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(s.URL)
}

func InitializeHealthCheckInterval(healthCheckInterval string) time.Duration {
	intervalOnce.Do(func() {
		var err error
		interval, err = time.ParseDuration(healthCheckInterval)
		if err != nil {
			interval = (time.Second * 2)
		}
	})

	return interval
}

func InitializeMaxUnhealthyChecks(count int8) {
	if count > 0 {
		maxUnhealthyChecks = count
	}
}
