package main

import (
	"sync"
)

type RoundRobinStrategy struct {
	Current int
	Mutex   sync.Mutex
}
