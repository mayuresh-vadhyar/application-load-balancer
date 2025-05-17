package main

type LoadBalancingStrategy interface {
	GetNextServer([]*Server) *Server
	CreateServerList(Config) []*Server
}
