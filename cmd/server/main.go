package main

import (
	"fmt"
	"log"
	"os"

	"github.com/antunesleo/cheflb/internal/lbs"
	"github.com/antunesleo/cheflb/internal/server"
)

const (
	envLayer     = "CHEFLB_LAYER"     // "4" or "7"
	envAlgorithm = "CHEFLB_ALGORITHM" // "round_robin", "hash", "least_response_time"
)

func main() {
	layer := getenv(envLayer, "7")
	algorithm := getenv(envAlgorithm, "round_robin")

	loadBalancer, err := buildLoadBalancer(algorithm)
	if err != nil {
		log.Fatalf("%s: %v", envAlgorithm, err)
	}

	log.Printf("starting cheflb on layer %s with %s balancing", layer, algorithm)

	switch layer {
	case "4":
		server.Layer4TcpStart(loadBalancer)
	case "7":
		server.Layer7HttpStart(loadBalancer)
	default:
		log.Fatalf("%s: invalid value %q (expected \"4\" or \"7\")", envLayer, layer)
	}
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func buildLoadBalancer(algorithm string) (lbs.LoadBalancer, error) {
	servers := lbs.NewServers()
	switch algorithm {
	case "round_robin":
		return lbs.NewRoundHobinLb(servers), nil
	case "hash":
		return lbs.NewHashLb(servers), nil
	case "least_response_time":
		return lbs.NewLeastRespTimeLb(servers), nil
	default:
		return nil, fmt.Errorf("invalid value %q (expected round_robin, hash, or least_response_time)", algorithm)
	}
}
