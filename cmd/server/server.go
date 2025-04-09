package server

import (
	"fmt"
	"http_protocole/cmd/response"
	"http_protocole/internal/request"
	"net"
	"sync/atomic"
)

type HandlerError struct {
	StatusCode int
	Msg        string
}

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	handler  Handler
	Listener net.Listener
	closed   atomic.Bool
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.Listener.Close()
}

func (s *Server) listen() {
	if s.closed.Load() {
		fmt.Printf("Connection closed\n")
		return
	}
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.closed.Load() {
				break
			}
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	w := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		w.WriteStatusLine(response.StatusCodeBadRequest)
		body := []byte(fmt.Sprintf("Error parsing request: %v", err))
		w.WriteHeaders(response.GetDefaultHeaders(len(body)))
		w.WriteBody(body)
		return
	}
	s.handler(w, req)
}

func Serve(port int, handler Handler) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return nil, err
	}
	server := Server{
		Listener: l,
		handler:  handler,
	}
	go server.listen()

	return &server, nil
}
