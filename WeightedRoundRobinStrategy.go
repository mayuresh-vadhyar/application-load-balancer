package main

import "sync"

type WeightedRoundRobinStrategy struct {
	Mutex sync.Mutex
}
