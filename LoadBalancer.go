package main

import (
	"log"
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
		log.Println(server.URL.String(), " is healthy -> ", server.IsHealthy)
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
		log.Println("Returning nil")
		return nil
	}
	nextServer.CurrentWeight -= totalWeight
	nextServer.Mutex.Unlock()
	log.Println("Returning server :", nextServer.URL.String())
	return nextServer
}
