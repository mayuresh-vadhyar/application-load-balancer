package main

import (
	"fmt"
	"net/http"
	"time"
)

func HealthCheck(s *Server, healthCheckInterval time.Duration) {
	for range time.Tick(healthCheckInterval) {
		res, err := http.Head(s.URL.String())
		s.Mutex.Lock()
		if err == nil && res.StatusCode == http.StatusOK {
			s.IsHealthy = true
		} else {
			fmt.Printf("%s is down\n", s.URL)
			s.IsHealthy = false
		}
		s.Mutex.Unlock()
	}
}
