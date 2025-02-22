package server

import (
	"fmt"
	"io"
	"net"

	"github.com/antunesleo/cheflb/internal/lbs"
)

func Layer4TcpStart() {
	fmt.Println("Welcome to Chef Loadbalancer!")
	fmt.Println("running it on layer4")

	servers := lbs.NewServers()
	loadBalancer := lbs.NewHashLb(servers)

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
	ipAddress := conn.LocalAddr().String()
	server := lb.Balance(ipAddress)

	remoteConn, err := net.Dial("tcp", server.UrlWithoutProtocolPrefix())
	if err != nil {
		fmt.Println("failed to connect to remote server", err, server.Url)
		// TODO: We need to return error to the tcp client, but not sure how
		return
	}
	defer remoteConn.Close()

	go io.Copy(remoteConn, conn)
	io.Copy(conn, remoteConn)
}
