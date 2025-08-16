package main

import (
	"log"
	"net/http"

	"github.com/mayuresh-vadhyar/application-load-balancer/constants"
	"github.com/mayuresh-vadhyar/application-load-balancer/server"
)

type LoadBalancingStrategy interface {
	GetNextServer(servers []*Server, r *http.Request) *Server
	CreateServerList(Config) []*Server
}

func CreateServerList(config Config) []*Server {
	serverUrls := config.Servers

	for _, rawUrl := range serverUrls {
		item, err := server.CreateServer(rawUrl)
		if err != nil {
			continue
		}
		server.Servers = append(server.Servers, item)
	}

	return server.Servers
}

func GetLoadBalancingStrategy(algorithm string) LoadBalancingStrategy {
	if algorithm == "" {
		algorithm = constants.ROUND_ROBIN
	}

	switch algorithm {
	case constants.ROUND_ROBIN:
		return &RoundRobinStrategy{Current: -1}
	case constants.WEIGHTED_ROUND_ROBIN:
		return &WeightedRoundRobinStrategy{Current: -1}
	case constants.IP_HASH:
		return &IPHashStrategy{}
	case constants.URL_HASH:
		return &URLHashStrategy{}
	default:
		log.Fatalf("Unknown algorithm: %s", algorithm)
	}

	return nil
}
