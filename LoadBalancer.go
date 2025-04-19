package LoadBalancer

type LoadBalancer struct {
    Current int
    Mutex   sync.Mutex
}

func (lb *LoadBalancer) GetNextServer(servers []*Servers) *Server {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()
	
	countOfServers := len(servers)
	for i:=0; i < countOfServers; i++ {
		lb.Current = (lb.Current + 1) % countOfServers
		nextServer = servers[i]

		nextServer.Mutex.Lock()
		isHealthy = nextServer.IsHealthy
		nextServer.Mutex.Unlock()

		if isHealthy {
			return nextServer
		}
	}

	return nil

}

http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	server := lb.GetNextServer(config.Servers)
	if (server == nil) {
		http.Error(w, "No healthy server available", http.StatusServiceUnavailable)
		return
	}

	w.Header().Add("X-Forwarded-Server", server.URL.String())
	server.ReverseProxy().ServeHTTP(w, r)
})