package server

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/antunesleo/cheflb/internal/lbs"
)

// "time"

const forwardMode = "request" // request or redirect

type LbHandler struct {
	loadBalancer lbs.LoadBalancer
}

func (mh *LbHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	fmt.Println("Request received!")

	targetServer := mh.loadBalancer.Balance(r.RemoteAddr)

	targetUrl := fmt.Sprintf("%s%s", targetServer.Url, r.URL.Path)
	
	if forwardMode == "redirect" {
		fmt.Println("redirect")
		rw.Header().Add(
			"Location", targetUrl,
		)
		rw.WriteHeader(307)
	} else {
		fmt.Println("forward")
		parsedUrl, err := url.Parse(targetUrl)
		if err != nil {
			return
		}
		rproxy := httputil.NewSingleHostReverseProxy(parsedUrl)
		
		startTime := time.Now()
		rproxy.ServeHTTP(rw, r)
		targetServer.UpdateMeanResponseTime(time.Since(startTime))
	}

}

func Layer7HttpStart() {
	fmt.Println("Welcome to Chef Loadbalancer!")

	servers := lbs.NewServers()
	loadBalancer := lbs.NewLeastRespTimeLb(servers)
	myHandler := &LbHandler{loadBalancer}
	server := &http.Server{
		Addr: ":8080",
		Handler: myHandler,
	}

	log.Fatal(server.ListenAndServe())
}
