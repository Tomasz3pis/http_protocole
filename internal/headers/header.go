package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

const crlf = "\r\n"
const invalidChars = ""

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		// the empty line
		// headers are done, consume the CRLF
		return 2, true, nil
	}

	parts := bytes.SplitN(data[:idx], []byte(":"), 2)
	key := string(parts[0])

	if key != strings.TrimRight(key, " ") {
		return 0, false, fmt.Errorf("invalid header name: %s", key)
	}

	value := bytes.TrimSpace(parts[1])
	key = strings.ToLower(strings.TrimSpace(key))
	if !validTokens([]byte(key)) {
		return 0, false, fmt.Errorf("invalid header token found: %s", key)
	}
	h.Set(key, string(value))
	return idx + 2, false, nil
}

var tokenChars = []byte{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}

// validTokens checks if the data contains only valid tokens
// or characters that are allowed in a token
func validTokens(data []byte) bool {
	for _, c := range data {
		if !(c >= 'A' && c <= 'Z' ||
			c >= 'a' && c <= 'z' ||
			c >= '0' && c <= '9' ||
			c == '-') {
			return false
		}
	}
	return true
}
func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	if val, ok := h[key]; ok {
		value = fmt.Sprintf("%s, %s", val, value)
	}
	h[key] = value
}

func (h Headers) Get(key string) (string, bool) {
	val, ok := h[strings.ToLower(key)]
	return val, ok
}

func (h Headers) Override(key, value string) {
	key = strings.ToLower(key)
	h[key] = value
}
