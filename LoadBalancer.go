package main

import (
	"sync"
)

type LoadBalancer struct {
	Current int
	Mutex   sync.Mutex
}

func (lb *LoadBalancer) GetNextServer(servers []*Server) *Server {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()

	countOfServers := len(servers)
	for i := 0; i < countOfServers; i++ {
		lb.Current = (lb.Current + 1) % countOfServers
		nextServer := servers[lb.Current]

		nextServer.Mutex.Lock()
		isHealthy := nextServer.IsHealthy
		nextServer.Mutex.Unlock()

		if isHealthy {
			return nextServer
		}
	}

	return nil

}

func (lb *LoadBalancer) GetNextServerForWRR(servers []*Server) *Server {
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
