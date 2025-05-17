package main

import "sync"

type WeightedRoundRobinStrategy struct {
	Current int
	Mutex   sync.Mutex
}

func (w *WeightedRoundRobinStrategy) CreateServerList(Config) []*Server {
	panic("unimplemented CreateServerList of WRR Strategy")
}

func (w *WeightedRoundRobinStrategy) GetNextServer([]*Server) *Server {
	panic("unimplemented GetNextServer of WRR Strategy")
}
