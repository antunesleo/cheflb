package server

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/antunesleo/cheflb/internal/lbs"
)

// Layer4TcpStart accepts TCP on :8080 and pipes bytes between each client
// and a backend picked once per connection by loadBalancer.
//
// L4 trade-offs: protocol-agnostic and fast (any protocol over TCP), but
// can't route by URL/header and can't rebalance requests sharing one
// keep-alive/HTTP-2 connection.
//
// Limits here: no timeouts, no health checks, no graceful shutdown. The
// "response time" recorded for least-response-time balancing is TCP
// session lifetime, not request latency — mostly noise on L4.
func Layer4TcpStart(loadBalancer lbs.LoadBalancer) {
	fmt.Println("Welcome to Chef Loadbalancer!")
	fmt.Println("running it on layer4")

	tcpListener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer tcpListener.Close()

	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
		}
		go handleConn(conn, loadBalancer)
	}

}

func handleConn(conn net.Conn, lb lbs.LoadBalancer) {
	defer conn.Close()
	clientIP, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		clientIP = conn.RemoteAddr().String()
	}
	server := lb.Balance(clientIP)
	fmt.Printf("balancing %s -> %s\n", clientIP, server.Url)

	remoteConn, err := net.Dial("tcp", server.UrlWithoutProtocolPrefix())
	if err != nil {
		fmt.Println("failed to connect to remote server", err, server.Url)
		// TODO: We need to return error to the tcp client, but not sure how
		return
	}
	defer remoteConn.Close()

	forwardConnectionData(remoteConn, conn, server)
}

func forwardConnectionData(remoteConn net.Conn, conn net.Conn, server *lbs.Server) {
	var wg sync.WaitGroup
	wg.Add(2)

	startTime := time.Now()
	
	go func() {
		defer wg.Done()
		io.Copy(remoteConn, conn)
		// Close the write side of the remote connection to signal EOF
		if c, ok := remoteConn.(*net.TCPConn); ok {
			c.CloseWrite()
		}
	}()
	
	go func() {
		defer wg.Done()
		io.Copy(conn, remoteConn)
		// Close the write side of the client connection to signal EOF
		if c, ok := conn.(*net.TCPConn); ok {
			c.CloseWrite()
		}
	}()

	// Wait for both copy operations to complete
	wg.Wait()

	server.UpdateMeanResponseTime(time.Since(startTime))
}
