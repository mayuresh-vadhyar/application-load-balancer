package main

import (
	"log"
	"net/http"
)

type LoadBalancingStrategy interface {
	GetNextServer(servers []*Server, r *http.Request) *Server
	CreateServerList(Config) []*Server
}

func GetLoadBalancingStrategy(algorithm string) LoadBalancingStrategy {
	if algorithm == "" {
		algorithm = "RR"
	}

	switch algorithm {
	case "RR":
		return &RoundRobinStrategy{Current: -1}
	case "WRR":
		return &WeightedRoundRobinStrategy{Current: -1}
	case "IPHash":
		return &IPHashStrategy{}
	default:
		log.Fatalf("Unknown algorithm: %s", algorithm)
	}

	return nil
}
