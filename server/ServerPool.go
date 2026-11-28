package server

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/mayuresh-vadhyar/application-load-balancer/Redis"
)

type ServerPoolServer struct {
	Id              int    `json:"id"`
	IsHealthy       bool   `json:"isHealthy"`
	Weight          int    `json:"weight"`
	CurrentWeight   int    `json:"currentWeight"`
	UnhealthyChecks int8   `json:"unhealthyChecks"`
	RequestCount    int64  `json:"requestCount"`
	ActiveReqCount  int64  `json:"activeRequestCount"`
	URL             string `json:"url"`
}

func getServerPoolKey(id string) string {
	return "SERVER_POOL:" + id
}

func ExistsInServerPool(id string) bool {
	client := Redis.GetClient()
	if client == nil {
		return false
	}
	return client.Exists(context.Background(), getServerPoolKey(id)).Val() > 0
}

func getServerDataFromPool(id string) string {
	client := Redis.GetClient()
	if client == nil {
		log.Printf("Redis client not initialized. Skipping Server Pool based initialization")
		return ""
	}

	ctx := context.Background()
	result := client.Get(ctx, getServerPoolKey(id))
	if result.Err() != nil {
		log.Printf("Could not find server pool logs in Redis for Load Balancer with ID: %s %s", id, result.Err())
		return ""
	}
	return result.Val()
}

func GetServersFromPool(id string) []*Server {
	data := getServerDataFromPool(id)
	serverList := []ServerPoolServer{}
	if err := json.Unmarshal([]byte(data), &serverList); err != nil && data != "" {
		log.Printf("Error unmarshaling server pool data for Load Balancer with ID: %s", id)
		return nil
	}

	servers := []*Server{}
	for _, serverObj := range serverList {
		server, recreateErr := RecreateServer(serverObj)
		if recreateErr != nil {
			log.Printf("Error recreating server: %v", recreateErr)
			return nil
		}
		servers = append(servers, server)
	}
	return servers
}

func StartServerPoolLogRoutine(ctx context.Context, config Config) <-chan struct{} {
	done := make(chan struct{})
	logInterval, parseErr := time.ParseDuration(config.ServerPoolInterval)
	if parseErr != nil || logInterval <= 0 {
		log.Printf("Server Pool Interval not configured. Skipping Server Pool Logging")
		close(done)
		return done
	}

	client := Redis.GetClient()
	if client == nil {
		log.Printf("Redis client not initialized. Skipping Server Pool Logging")
		close(done)
		return done
	}

	ticker := time.NewTicker(logInterval)
	id := config.Id
	expiry, expiryErr := time.ParseDuration(config.ServerPoolExpiry)
	if expiryErr != nil || expiry <= 0 {
		expiry = (time.Hour * 2)
	}

	go func() {
		defer close(done)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Printf("Stopping Server Pool Log Routine for Load Balancer with ID: %s", id)
				return
			case <-ticker.C:
				poolServers := make([]ServerPoolServer, 0, len(GetServers()))
				for _, server := range GetServers() {
					poolServers = append(poolServers, ServerPoolServer{
						Id:              server.Id,
						IsHealthy:       server.IsHealthy,
						Weight:          server.Weight,
						UnhealthyChecks: server.UnhealthyChecks,
						RequestCount:    server.RequestCount,
						ActiveReqCount:  server.ActiveReqCount,
						URL:             server.URL.Host,
					})
				}

				servers, err := json.Marshal(poolServers)
				if err != nil {
					log.Printf("Error parsing servers for server pool: %v", err)
					continue
				}

				redisErr := client.Set(ctx, getServerPoolKey(id), servers, expiry).Err()
				if redisErr != nil {
					log.Printf("SERVER POOL LOGGING ERROR: %v", redisErr)
				}
			}
		}
	}()
	return done
}
