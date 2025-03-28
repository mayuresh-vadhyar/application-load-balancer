type LoadBalancer struct {
    Current int
    Mutex   sync.Mutex
}

type Server struct {
    URL       *url.URL
    IsHealthy bool
    Mutex     sync.Mutex
}

func (lb *LoadBalancer) getNextServer(servers []*Servers) *Server {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()
	
	countOfServers := len(servers)
	for i:=0; i < countOfServers; i++ {
		lb.Current = (lb.Current + 1) % countOfServers
		nextServer = servers[idx]

		nextServer.Mutex.Lock
		isHealthy = nextServer.IsHealthy
		nextServer.Mutex.Unlock

		if(isHealthy) {
			return nextServer
		}
	}

	return nil

}