package main

import (
	"crypto/sha256"
	"fmt"
	"http_protocole/cmd/response"
	"http_protocole/cmd/server"
	"http_protocole/internal/headers"
	"http_protocole/internal/request"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		handler400(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, req)
		return
	}
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		handlerProxy(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/video" && req.RequestLine.Method == "GET" {
		handleVideo(w, req)
		return
	}
	handler200(w, req)
}

func handler400(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeBadRequest)
	body := []byte(`<html>
<head>
<title>400 Bad Request</title>
</head>
<body>
<h1>Bad Request</h1>
<p>Your request honestly kinda sucked.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler500(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeInternalServerError)
	body := []byte(`<html>
<head>
<title>500 Internal Server Error</title>
</head>
<body>
<h1>Internal Server Error</h1>
<p>Okay, you know what? This one is on me.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler200(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeSuccess)
	body := []byte(`<html>
<head>
<title>200 OK</title>
</head>
<body>
<h1>Success!</h1>
<p>Your request was an absolute banger.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handlerProxy(w *response.Writer, req *request.Request) {
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	resp, err := http.Get(fmt.Sprintf("https://httpbin.org/%s", target))
	if err != nil {
		handler500(w, req)
		return
	}
	defer resp.Body.Close()
	w.WriteStatusLine(response.StatusCodeSuccess)
	h := response.GetDefaultHeaders(0)
	delete(h, "content-lenght")
	h.Override("Content-Type", "text/html")
	h.Set("Transfer-Encoding", "chunked")
	h.Set("Trailer", "X-Content-SHA256")
	h.Set("Trailer", "X-Content-Length")
	w.WriteHeaders(h)
	fullBody := make([]byte, 0)
	buffer := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buffer)
		fmt.Println("Read", n, "bytes")
		if n > 0 {
			_, err = w.WriteChunkedBody(buffer[:n])
			if err != nil {
				fmt.Println("Error writing chunked body:", err)
				break
			}
			fullBody = append(fullBody, buffer[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading response body:", err)
			break
		}
	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Println("Error writing chunked body done:", err)
	}
	sha256 := fmt.Sprintf("%x", sha256.Sum256(fullBody))
	t := headers.NewHeaders()
	t.Set("X-Content-SHA256", sha256)
	t.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
	err = w.WriteTrailers(t)
	if err != nil {
		fmt.Println("Error writing trailers:", err)
	}
	fmt.Println("Wrote trailers")
}

func handleVideo(w *response.Writer, req *request.Request) {
	w.WriteStatusLine(response.StatusCodeSuccess)
	body, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		handler500(w, req)
		return
	}
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "video/mp4")
	err = w.WriteHeaders(h)
	if err != nil {
		fmt.Println("Error writing headers:", err)
	}
	_, err = w.WriteBody(body)
	if err != nil {
		fmt.Println("Error writing body:", err)
	}
}
