package lbs


type Server struct {}

type LoadBalancer interface {
	Balance() Server
}

type RoundHobinLb struct {
	servers []*Server
	index int	
}

func (lb *RoundHobinLb) Balance() *Server {
	serverToReturn := lb.servers[lb.index]
	if lb.index == len(lb.servers) {
		lb.index = 0
	} else {
		lb.index += 1
	}

	return serverToReturn
}

func NewRoundHobinLb(servers []*Server) *RoundHobinLb {
	return &RoundHobinLb{servers, 0}
}
