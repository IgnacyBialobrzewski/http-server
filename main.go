package main

import (
	"http-server/http"
	"log"
)

func main() {
	s := http.Server{}

	s.HandleRequest(func(req *http.HttpRequestMessage) {
		log.Println("Wow I'm handling a request")
		log.Printf("%s\n", req.Method)
	})

	if err := s.Start(":8080"); err != nil {
		panic(err)
	}
}
