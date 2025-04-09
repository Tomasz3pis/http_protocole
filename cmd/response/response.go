package response

import (
	"fmt"
	"http_protocole/internal/headers"
	"io"
)

type responseState int

const (
	Initialize responseState = iota
	StatusLineWritten
	HeadersWritten
	BodyWritten
	TrailersWritten
)

type Writer struct {
	Writer io.Writer
	State  responseState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		Writer: w,
		State:  Initialize,
	}
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.State != BodyWritten {
		return fmt.Errorf("cannot write trailers in state %d", w.State)
	}
	defer func() { w.State = TrailersWritten }()
	for k, v := range h {
		_, err := w.Writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
		if err != nil {
			return err
		}
	}
	_, err := w.Writer.Write([]byte("\r\n"))
	return err
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.State != Initialize {
		return fmt.Errorf("Writer in an incorrect state: %v", w.State)
	}
	statusLine := getStatusLine(statusCode)
	_, err := w.Writer.Write(statusLine)
	if err != nil {
		return err
	}
	w.State = StatusLineWritten
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.State != StatusLineWritten {
		return fmt.Errorf("Writer in an incorrect state: %v", w.State)
	}
	for k, v := range headers {
		_, err := w.Writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
		if err != nil {
			return err
		}
	}
	_, err := w.Writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	w.State = HeadersWritten
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.State != HeadersWritten {
		return 0, fmt.Errorf("Writer in an incorrect state: %v", w.State)
	}
	n, err := w.Writer.Write(p)
	if err != nil {
		return 0, err
	}
	w.State = BodyWritten
	return n, nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	chunkSizeHex := fmt.Sprintf("%x", len(p))
	n, err := w.Writer.Write([]byte(chunkSizeHex + "\r\n"))
	if err != nil {
		return n, err
	}
	n, err = w.Writer.Write(p)
	if err != nil {
		return n, err
	}
	n, err = w.Writer.Write([]byte("\r\n"))
	if err != nil {
		return n, err
	}

	return len(p), nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	w.State = BodyWritten
	n, err := w.Writer.Write([]byte("0\r\n"))
	return n, err
}

type StatusCode int

const (
	StatusCodeSuccess             StatusCode = 200
	StatusCodeBadRequest          StatusCode = 400
	StatusCodeInternalServerError StatusCode = 500
)

func getStatusLine(statusCode StatusCode) []byte {
	reasonPhrase := ""
	switch statusCode {
	case StatusCodeSuccess:
		reasonPhrase = "OK"
	case StatusCodeBadRequest:
		reasonPhrase = "Bad Request"
	case StatusCodeInternalServerError:
		reasonPhrase = "Internal Server Error"
	}
	return []byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase))
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}
