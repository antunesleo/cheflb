package server

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/antunesleo/cheflb/internal/lbs"
)

// "time"

const forwardMode = "redirect" // request or redirect

type LbHandler struct {
	loadBalancer lbs.LoadBalancer
}

func (mh *LbHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	fmt.Println("Request received!")

	targetServer := mh.loadBalancer.Balance(r.RemoteAddr)

	url := fmt.Sprintf("%s%s", targetServer.Url, r.URL.Path)
	
	if forwardMode == "redirect" {
		fmt.Println("redirect")
		rw.Header().Add(
			"Location", url,
		)
		rw.WriteHeader(307)
	} else {
		fmt.Println("forward")

		client := &http.Client{}
		var reqBuffer []byte
		_, err := r.Body.Read(reqBuffer)
		if err != nil {
			return
		}
		
	
		req, err := http.NewRequest(r.Method, url, bytes.NewReader(reqBuffer))
		if err != nil {
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			return
		}

		var respBuffer []byte
		_, err = resp.Body.Read(respBuffer)
		if err != nil {
			return
		}
		rw.Write(respBuffer)
		rw.WriteHeader(resp.StatusCode)

	}

}

func Start() {
	fmt.Println("Welcome to Chef Loadbalancer!")

	servers := []*lbs.Server{
		lbs.NewServer("http://localhost:7171"),
		lbs.NewServer("http://localhost:8181"),
	}
	loadBalancer := lbs.NewHashLb(servers)
	myHandler := &LbHandler{loadBalancer}
	server := &http.Server{
		Addr: ":8080",
		Handler: myHandler,
	}

	log.Fatal(server.ListenAndServe())
}
