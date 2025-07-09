package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

func StartHealthCheckRoutine(ctx context.Context, s *Server, healthCheckInterval time.Duration) {
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Stopping health check for %s\n", s.URL.String())
			return

		case <-ticker.C:
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
}
