package main

import (
	"hash/fnv"
	"net/http"
	"sync"
)

type URLHashStrategy struct {
	Mutex sync.Mutex
}

func (lb *URLHashStrategy) CreateServerList(config Config) []*Server {
	var servers []*Server
	serverUrls := config.Servers

	for _, rawUrl := range serverUrls {
		server, err := CreateServer(rawUrl)
		if err != nil {
			continue
		}
		servers = append(servers, server)
	}

	return servers

}

func (lb *URLHashStrategy) GetNextServer(servers []*Server, r *http.Request) *Server {
	url := r.URL.Path
	n := len(servers)
	hash := fnv.New32a()
	hash.Write([]byte(url))
	i := int(hash.Sum32()) % n

	startIndex := i

	for {
		server := servers[i]
		server.Mutex.Lock()
		isHealthy := server.IsHealthy
		server.Mutex.Unlock()

		if isHealthy {
			return server
		}

		i = (i + 1) % n
		if i == startIndex {
			break
		}
	}

	return nil
}
