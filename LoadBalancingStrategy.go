package main

type LoadBalancingStrategy interface {
	GetNextServer([]*Server) *Server
	CreateServerList([]string) []*Server
}
