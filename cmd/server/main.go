package main

import (
	"github.com/antunesleo/cheflb/internal/server"
)
const layer = 7 // options: 4 or 7

func main() {
	if layer == 4 {
		server.Layer4TcpStart()
	} else if layer == 7 {
		server.Layer7HttpStart()
	}
}
