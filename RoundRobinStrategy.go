package main

import (
	"sync"
)

type RoundRobinStrategy struct {
	Current int
	Mutex   sync.Mutex
}

func (r *RoundRobinStrategy) CreateServerList(Config) []*Server {
	panic("unimplemented CreateServerList of RR Strategy")
}

func (r *RoundRobinStrategy) GetNextServer([]*Server) *Server {
	panic("unimplemented GetNextServer of RR Strategy")
}
