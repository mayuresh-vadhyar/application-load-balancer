package main

import (
	"net/http"
	"sync"

	"github.com/mayuresh-vadhyar/application-load-balancer/server"
)

type WeightedRoundRobinStrategy struct {
	Current int
	Mutex   sync.Mutex
}

func (lb *WeightedRoundRobinStrategy) CreateServerList(config Config) []*Server {
	serverUrls := config.Servers
	weights := config.Weights

	countOfServers := len(serverUrls)

	if weights == nil {
		weights := make([]int, countOfServers)
		for i := range weights {
			weights[i] = 1
		}
	}
	for i := 0; i < countOfServers; i++ {
		item, err := server.CreateWeightedServer(serverUrls[i], weights[i])
		if err != nil {
			continue
		}
		server.Servers = append(server.Servers, item)
	}

	return server.Servers
}

func (lb *WeightedRoundRobinStrategy) GetNextServer(servers []*Server, r *http.Request) *Server {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()

	var totalWeight int
	var nextServer *Server

	for _, server := range servers {
		server.Mutex.Lock()
		if !server.IsHealthy {
			server.Mutex.Unlock()
			continue
		}

		server.CurrentWeight += server.Weight
		totalWeight += server.Weight

		if nextServer == nil || server.CurrentWeight > nextServer.CurrentWeight {
			if nextServer != nil {
				nextServer.Mutex.Unlock()
			}

			nextServer = server
		} else {
			server.Mutex.Unlock()
		}
	}

	if nextServer == nil {
		return nil
	}
	nextServer.CurrentWeight -= totalWeight
	nextServer.Mutex.Unlock()
	return nextServer
}
