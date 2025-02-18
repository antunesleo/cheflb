package lbs

import (
	"sync"
	"github.com/spaolacci/murmur3"
)


type Server struct {
	Url string
}

func NewServer(url string) *Server {
	return &Server{url}
}

type LoadBalancer interface {
	Balance(ipAddress string) *Server
}

type RoundRobinLb struct {
	servers []*Server
	index int
	mu sync.Mutex
}

func (lb *RoundRobinLb) Balance(ipAddress string) *Server {
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

type HashLb struct {
	servers []*Server
}

func (lb *HashLb) Balance(ipAddress string) *Server {
	hash := murmur3.Sum32([]byte("example"))
	index := hash % uint32(len(lb.servers))
	return lb.servers[index]
}

func NewHashLb(servers []*Server) *HashLb {
	return &HashLb{servers: servers}
}
