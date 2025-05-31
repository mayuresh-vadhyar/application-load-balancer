package main

import (
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
			URL: 	   parsedUrl,
			IsHealthy: true
		}

		servers = append(servers, server)
	}

	return servers

}

func (lb *URLHashStrategy) GetNextServer(servers []*Server, r *http.Request) *Server {
	panic("unimplemented GetNextServer of URL Hash Strategy")
}
