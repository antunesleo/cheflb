package server

import (
	"fmt"
	"log"
	"net/http"
)

type LbHandler struct {}

func (mh *LbHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	fmt.Println("Request received!")
	url := fmt.Sprintf("http://localhost:7171%s", r.URL.Path)
	rw.Header().Add(
		"Location", url,
	)
	rw.WriteHeader(307)
}

func Start() {
	fmt.Println("Welcome to Chef Loadbalancer!")

	myHandler := &LbHandler{}
	server := &http.Server{
		Addr: ":8080",
		Handler: myHandler,
	}

	log.Fatal(server.ListenAndServe())
}
