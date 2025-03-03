package lbs

import (
	"strings"
	"sync"
	"time"

	"github.com/spaolacci/murmur3"
)


type Server struct {
	Url string
	AvgResponseTime int // miliseconds
}

func (s *Server) UpdateMeanResponseTime(responseTimeDuration time.Duration) {
	s.AvgResponseTime = (s.AvgResponseTime + int(responseTimeDuration.Milliseconds())) / 2
}

func (s *Server) UrlWithoutProtocolPrefix() string {
	prefix := "http://"
	if strings.HasPrefix(prefix, prefix) {
		return s.Url[len(prefix):]
	}
	return s.Url
}

func NewServer(url string) *Server {
	return &Server{url, 0}
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

type LeastRespTimeLb struct {
	servers []*Server
}

func (lb *LeastRespTimeLb) Balance(ipAddress string) *Server {
	var leastServer *Server
	for _, server := range lb.servers {
		if leastServer == nil {
			leastServer = server
		} else if server.AvgResponseTime < leastServer.AvgResponseTime {
			leastServer = server
		}
	}
	return leastServer
}

func NewLeastRespTimeLb(servers []*Server) *LeastRespTimeLb {
	return &LeastRespTimeLb{servers: servers}
}

func NewServers() []*Server{
	return []*Server{
		NewServer("http://localhost:7171"),
		NewServer("http://localhost:8181"),
	}
}