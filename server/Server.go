package server

import (
	"context"
	"encoding/json"
	"net/http/httputil"
	"net/url"
	"slices"
	"sync"
	"time"

	"github.com/mayuresh-vadhyar/application-load-balancer/config"
)

type Server struct {
	Id              int                `json:"id"`
	URL             *url.URL           `json:"-"`
	IsHealthy       bool               `json:"isHealthy"`
	Weight          int                `json:"weight,omitempty"`
	CurrentWeight   int                `json:"-"`
	Mutex           sync.Mutex         `json:"-"`
	StopHealthCheck context.CancelFunc `json:"-"`
	UnhealthyChecks int8               `json:"unhealthyChecks"`
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

type HealthCheckConfig = config.HealthCheckConfig

var Servers []*Server
var interval time.Duration
var cooldown time.Duration
var maxRestart int8
var healthCheckOnce sync.Once
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
	go StartHealthCheckRoutine(ctx, server, maxRestart)

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
	go StartHealthCheckRoutine(ctx, server, maxRestart)

	return server, nil
}

func (s *Server) ReverseProxy() *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(s.URL)
}

func InitializeHealthCheckConfig(healthCheckConfig HealthCheckConfig) {
	healthCheckOnce.Do(func() {
		var intervalErr error
		var cooldownErr error
		interval, intervalErr = time.ParseDuration(healthCheckConfig.Interval)
		if intervalErr != nil || interval <= 0 {
			interval = (time.Second * 2)
		}

		maxRestart = healthCheckConfig.MaxRestart
		cooldown, cooldownErr = time.ParseDuration(healthCheckConfig.Cooldown)
		if cooldownErr != nil || cooldown <= 0 || cooldown <= interval {
			cooldown = 0
			maxRestart = 0
		}

		if healthCheckConfig.MaxUnhealthyChecks > 0 {
			maxUnhealthyChecks = healthCheckConfig.MaxUnhealthyChecks
		}
	})
}
