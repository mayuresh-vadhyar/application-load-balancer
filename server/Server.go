package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"slices"
	"sync"
	"time"

	"github.com/mayuresh-vadhyar/application-load-balancer/Redis"
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
type Config = config.Config

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
	// TODO: Cache hot requests
	proxy := httputil.NewSingleHostReverseProxy(s.URL)

	// Case 1: Upstream responds with a 502
	proxy.ModifyResponse = func(r *http.Response) error {
		if r.StatusCode == http.StatusBadGateway {
			s.markUnhealthy()
		}
		return nil
	}

	// Case 2: Proxy cannot reach upstream (timeouts, connection errors, etc.)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		s.markUnhealthy()
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
	}
	return proxy
}

func (s *Server) markUnhealthy() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.UnhealthyChecks++
	s.IsHealthy = false
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

func StartServerPoolLogRoutine(config Config) {
	logInterval, parseErr := time.ParseDuration(config.ServerPoolInterval)
	if parseErr != nil || logInterval <= 0 {
		log.Printf("Server Pool Interval not configured. Skipping Server Pool Logging")
		return
	}

	client := Redis.GetClient()
	if client == nil {
		log.Printf("Redis client not initialized. Skipping Server Pool Logging")
		return
	}

	ctx := context.Background()
	ticker := time.NewTicker(logInterval)
	id := config.Id
	expiry, expiryErr := time.ParseDuration(config.ServerPoolExpiry)
	if expiryErr != nil || expiry <= 0 {
		expiry = (time.Hour * 2)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Printf("Stopping Server Pool Log Routine for Load Balancer with ID: %s", id)
				ticker.Stop()
				return
			case <-ticker.C:
				servers, err := json.Marshal(Servers)
				if err != nil {
					log.Printf("Error parsing servers for server pool: %v", err)
				}

				redisErr := client.Set(context.Background(), "SERVER_POOL:"+id, servers, expiry).Err()
				if redisErr != nil {
					log.Printf("SERVER POOL LOGGING ERROR: %v", redisErr)
				}
			}
		}
	}()
}
