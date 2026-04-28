package lbs

import (
	"strings"
	"sync/atomic"
	"time"

	"github.com/spaolacci/murmur3"
)


type Server struct {
	Url string
	// AvgResponseTime is read by Balance() and written by
	// UpdateMeanResponseTime() from many goroutines, so it's atomic. We tolerate
	// the rare lost-update from racing read-modify-writes — the metric is
	// already a noisy estimate — but reads must never tear or be optimized away.
	AvgResponseTime atomic.Int64 // milliseconds
}

func (s *Server) UpdateMeanResponseTime(responseTimeDuration time.Duration) {
	old := s.AvgResponseTime.Load()
	s.AvgResponseTime.Store((old + responseTimeDuration.Milliseconds()) / 2)
}

func (s *Server) UrlWithoutProtocolPrefix() string {
	prefix := "http://"
	if strings.HasPrefix(prefix, prefix) {
		return s.Url[len(prefix):]
	}
	return s.Url
}

func NewServer(url string) *Server {
	return &Server{Url: url}
}

type LoadBalancer interface {
	Balance(ipAddress string) *Server
}

type RoundRobinLb struct {
	servers []*Server
	// counter is incremented atomically per request; the index is derived
	// with `% len(servers)`. uint64 won't realistically overflow.
	counter atomic.Uint64
}

func (lb *RoundRobinLb) Balance(ipAddress string) *Server {
	i := lb.counter.Add(1) - 1
	return lb.servers[i%uint64(len(lb.servers))]
}

func NewRoundHobinLb(servers []*Server) *RoundRobinLb {
	return &RoundRobinLb{servers: servers}
}

// HashLb maps each client to a backend by hashing a client identifier and
// taking `hash % N`. Same client -> same server every time, with no state
// stored on the LB. Useful when servers hold per-client state (in-memory
// sessions, caches, WebSockets).
//
// Two production caveats this implementation does NOT address:
//
// 1) `hash % N` reshuffles almost everyone when the pool changes. Adding
//    or removing one server changes the modulo for most clients, so
//    sticky sessions break and per-client caches all miss at once.
//    Production LBs use algorithms that move only ~1/N of clients on a
//    pool change: consistent hashing (servers placed on a virtual ring,
//    each client maps to the next one clockwise) or rendezvous/HRW hashing
//    (hash (client, server) for every server, pick the highest score).
//
// 2) Hashing the client IP is fragile. Home wifi, office networks, and
//    mobile carriers all hide many devices behind a single public IP, so
//    those users collapse onto one backend and skew the load. A user's IP
//    also changes when they switch networks (wifi <-> LTE, VPN), so they
//    lose their assigned server. Real LBs usually hash a session cookie or
//    user ID instead — those actually identify the client, rather than
//    the network they happen to be on.
type HashLb struct {
	servers []*Server
}

func (lb *HashLb) Balance(ipAddress string) *Server {
	hash := murmur3.Sum32([]byte(ipAddress))
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
		} else if server.AvgResponseTime.Load() < leastServer.AvgResponseTime.Load() {
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