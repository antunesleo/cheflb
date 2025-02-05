package server

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	// "time"
)

const forwardMode = "request" // request or redirect

type LbHandler struct {}

func (mh *LbHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	fmt.Println("Request received!")
	url := fmt.Sprintf("http://localhost:7171%s", r.URL.Path)
	
	if forwardMode == "redirect" {
		rw.Header().Add(
			"Location", url,
		)
		rw.WriteHeader(307)
	} else {
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

	myHandler := &LbHandler{}
	server := &http.Server{
		Addr: ":8080",
		Handler: myHandler,
	}

	log.Fatal(server.ListenAndServe())
}
