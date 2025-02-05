package lbs

import "sync"


type Server struct {
	Url string
}

func NewServer(url string) *Server {
	return &Server{url}
}

type LoadBalancer interface {
	Balance() *Server
}

type RoundRobinLb struct {
	servers []*Server
	index int
	mu sync.Mutex
}

func (lb *RoundRobinLb) Balance() *Server {
	lb.mu.Lock()
	serverToReturn := lb.servers[lb.index]
	if lb.index == len(lb.servers)-1 {
		lb.index = 0
	} else {
		lb.index += 1
	}

	lb.mu.Unlock()
	return serverToReturn
}

func NewRoundHobinLb(servers []*Server) *RoundRobinLb {
	return &RoundRobinLb{servers: servers, index: 0}
}
