package main

import (
	"sync"
)

type IPHashStrategy struct {
	Mutex sync.Mutex
}

func (lb *IPHashStrategy) CreateServerList(config Config) []*Server {
	panic("unimplemented CreateServerList of IP Hash Strategy")
}

func (lb *IPHashStrategy) GetNextServer(servers []*Server) *Server {
	panic("unimplemented GetNextServer of IP Hash Strategy")
}
