package server

import (
	"context"
	"log"
	"net/http"
	"time"
)

func StartHealthCheckRoutine(ctx context.Context, s *Server, maxRestart int8) {
	go func() {
		var restartCount int8 = 0

		for {
			runHealthCheck(ctx, s, interval)

			if ctx.Err() != nil {
				log.Printf("Health check stopped for %s (context canceled)", s.URL)
				return
			}

			if cooldown <= 0 {
				log.Printf("Cooldown not configured. Not restarting health check for %s.", s.URL)
				return
			}

			if maxRestart > 0 && restartCount >= maxRestart {
				log.Printf("Max restarts reached for %s. Not restarting health check.", s.URL)
				return
			}

			select {
			case <-ctx.Done():
				log.Printf("Aborted cooldown restart for %s (context canceled)", s.URL)
				return
			case <-time.After(cooldown):
				restartCount++
				log.Printf("Restarting health check for %s after cooldown %v...", s.URL, cooldown)
			}
		}
	}()
}

func runHealthCheck(ctx context.Context, s *Server, healthCheckInterval time.Duration) {
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Stopping health check for %s\n", s.URL.String())
			return

		case <-ticker.C:
			log.Printf("Check health for %s\n", s.URL.String())
			res, err := http.Head(s.URL.String())
			s.Mutex.Lock()
			if err == nil && res.StatusCode == http.StatusOK {
				s.IsHealthy = true
				s.UnhealthyChecks = 0
			} else {
				log.Printf("%s is down\n", s.URL)
				if maxUnhealthyChecks > 0 && s.UnhealthyChecks+1 >= maxUnhealthyChecks {
					DeleteServer(s.URL.String())
				} else {
					s.UnhealthyChecks++
					s.IsHealthy = false
				}
			}
			s.Mutex.Unlock()
		}
	}
}
