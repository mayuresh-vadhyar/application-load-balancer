package main

import (
	"sync"
)

type URLHashStrategy struct {
	Mutex sync.Mutex
}

func (lb *URLHashStrategy) CreateServerList(config Config) []*Server {
	panic("unimplemented CreateServerList of URL Hash Strategy")
}

func (lb *URLHashStrategy) GetNextServer(servers []*Server) *Server {
	panic("unimplemented GetNextServer of URL Hash Strategy")
}
