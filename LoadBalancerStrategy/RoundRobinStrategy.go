package main

import (
	"net/http"
	"sync"
)

type RoundRobinStrategy struct {
	Current int
	Mutex   sync.Mutex
}

func (lbs *RoundRobinStrategy) CreateServerList(config Config) []*Server {
	return CreateServerList(config)
}

func (lbs *RoundRobinStrategy) GetNextServer(servers []*Server, _ *http.Request) *Server {
	lbs.Mutex.Lock()
	defer lbs.Mutex.Unlock()

	countOfServers := len(servers)
	for i := 0; i < countOfServers; i++ {
		lbs.Current = (lbs.Current + 1) % countOfServers
		nextServer := servers[lbs.Current]

		nextServer.Mutex.Lock()
		isHealthy := nextServer.IsHealthy
		nextServer.Mutex.Unlock()

		if isHealthy {
			return nextServer
		}
	}

	return nil
}
