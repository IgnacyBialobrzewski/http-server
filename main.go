package main

import (
	"log"

	"http1.1-server/pkg/http1"
)

func main() {
	s := http1.Server{}

	s.HandleRequest(func(req *http1.HttpRequestMessage) {
		log.Println("Wow I'm handling a request")
		log.Printf("%s\n", req.Method)
	})

	if err := s.Start(":8080"); err != nil {
		panic(err)
	}
}
