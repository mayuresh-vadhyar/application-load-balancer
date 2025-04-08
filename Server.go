package LoadBalancer

type Server struct {
    URL       *url.URL
    IsHealthy bool
    Mutex     sync.Mutex
}

func (s *Server) ReverseProxy() *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(s.URL)
}

log.Println("Starting load balancer on port", config.Port)
err = http.ListenAndServe(config.Port, nil)
if err != nil {
        log.Fatalf("Error starting load balancer: %s\n", err.Error())
}
