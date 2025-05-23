package main

import "log"

type LoadBalancingStrategy interface {
	GetNextServer([]*Server) *Server
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
	default:
		log.Fatalf("Unknown algorithm: %s", algorithm)
	}

	return nil
}
