package main

import (
	"fmt"
	"http_protocole/internal/request"
	"log"
	"net"
)

const port = ":42069"

func main() {

	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("error listening for TCP traffic: %s\n", err.Error())
	}
	defer l.Close()

	fmt.Println("Listening for TCP traffic on", port)
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalf("error: %s\n", err.Error())
		}
		fmt.Println("Accepted connection from", conn.RemoteAddr())

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("error reading request: %s", err)
		}
		fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)
		fmt.Printf("Headers:\n")
		for k, v := range req.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}
		fmt.Printf("Body:\n%s\n", string(req.Body))
		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
	}
}
