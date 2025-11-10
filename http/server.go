package http

import (
	"io"
	"log"
	"net"
	"time"
)

const chunkSize = 8192
const readTimeOut = time.Second * 5

type Server struct {
	handlers []func(*HttpRequestMessage)
}

func (s *Server) HandleRequest(handler func(*HttpRequestMessage)) {
	s.handlers = append(s.handlers, handler)
}

func (s *Server) Start(address string) error {
	ln, err := net.Listen("tcp4", address)

	if err != nil {
		return err
	}

	log.Printf("started %s\n", ln.Addr())

	for {
		conn, err := ln.Accept()

		if err != nil {
			log.Println(err)
		}

		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	var buf []byte

	start := time.Now()

	for {
		if err := conn.SetReadDeadline(time.Now().Add(readTimeOut)); err != nil {
			log.Println("couldn't set read deadline")
			break
		}

		chunk := make([]byte, chunkSize)
		n, err := conn.Read(chunk)

		if err != nil && err != io.EOF && n <= 0 {
			log.Println(err)
			break
		}

		buf = append(buf, chunk[:n]...)
		msg, err := ParseRequest(buf)

		if err != nil {
			if err == ErrBodyTooSmall {
				continue
			} else {
				log.Println(err)
				return
			}
		}

		log.Printf("request processed in %s\n", time.Since(start))

		for _, handler := range s.handlers {
			go func() {
				handler(&msg)
			}()
		}

		break
	}
}
