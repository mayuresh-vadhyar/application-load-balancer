package main

import (
	"hash/fnv"
	"net/http"
	"net/url"
	"sync"
)

type URLHashStrategy struct {
	Mutex sync.Mutex
}

func (lb *URLHashStrategy) CreateServerList(config Config) []*Server {
	var servers []*Server
	serverUrls := config.Servers

	for _, rawUrl := range serverUrls {
		parsedUrl, err := url.Parse(rawUrl)
		if err != nil {
			continue
		}

		server := &Server{
			URL:       parsedUrl,
			IsHealthy: true,
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
