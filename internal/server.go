package server

import (
	"fmt"
	"log"
	"net/http"
)

type MyHandler struct {}

func (mh *MyHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	fmt.Println("Request received!")
}

func Start() {
	fmt.Println("Welcome to Chef Loadbalancer!")

	myHandler := &MyHandler{}
	server := &http.Server{
		Addr: ":8080",
		Handler: myHandler,
	}

	log.Fatal(server.ListenAndServe())
}
