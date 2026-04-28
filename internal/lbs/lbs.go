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

// alpha is the EWMA weight on the newest sample. Higher = more reactive to
// recent changes, lower = smoother. 0.5 is quite reactive — one slow sample
// swings the average heavily; production EWMAs commonly use ~0.1 for
// stability.
const alpha = 0.5

// UpdateMeanResponseTime folds a new response-time sample into AvgResponseTime
// using an exponentially weighted moving average (EWMA).
//
// An EWMA is a running average where recent samples count more than older
// ones, with each sample's influence decaying exponentially as newer samples
// arrive. It's the standard alternative to a simple mean (which weights all
// samples equally forever) and a sliding window (which needs you to keep the
// last N values). The formula is:
//
//	new_avg = α·sample + (1−α)·old_avg
//
// Two nice properties: only one running number needs to be stored (no
// history buffer), and an old sample's weight after n updates is (1−α)^n —
// recent values dominate, ancient ones fade out automatically.
func (s *Server) UpdateMeanResponseTime(responseTimeDuration time.Duration) {
	old := s.AvgResponseTime.Load()
	sample := responseTimeDuration.Milliseconds()
	updated := int64(alpha*float64(sample) + (1-alpha)*float64(old))
	s.AvgResponseTime.Store(updated)
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

// RoundRobinLb cycles through servers in order: 0, 1, ..., N-1, 0, ....
// An atomic counter gives each request a unique ticket; `% N` picks the
// server. Fair under uniform load, no mutex needed.
//
// Trade-offs: no client affinity (one user's requests can hit different
// backends — bad for sticky sessions), load-blind (counts requests, not
// work — slow servers still get their turn), health-blind (a dead backend
// keeps being picked until removed from the slice).
type RoundRobinLb struct {
	servers []*Server
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

// LeastRespTimeLb picks the server with the lowest avg response time —
// adaptive, routes around slow backends without explicit health checks.
//
// Caveats:
//   - Cold start: all servers begin at 0, so traffic piles on servers[0]
//     until its avg rises, then onto servers[1]. P2C (pick 2 random, take
//     the lower) is the usual fix.
//   - The EWMA uses α=0.5, weighting the newest sample at 50%; one slow
//     request blackholes a server. Production EWMAs use α≈0.1.
//   - Latency != health: a fast 500 looks faster than a healthy 50ms.
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