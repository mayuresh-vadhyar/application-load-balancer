package main

import (
	"hash/fnv"
	"log"
	"net"
	"net/http"
	"sync"
)

type IPHashStrategy struct {
	Mutex sync.Mutex
}

func (lb *IPHashStrategy) CreateServerList(config Config) []*Server {
	panic("unimplemented CreateServerList of IP Hash Strategy")
}

func (lb *IPHashStrategy) GetNextServer(servers []*Server, r *http.Request) *Server {
	n := len(servers)
	clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Print("Failed to parse client IP from: ", r.RemoteAddr)
		return nil
	}

	hash := fnv.New32a()
	hash.Write([]byte(clientIP))
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
