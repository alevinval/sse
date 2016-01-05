package sse

import (
	"net/http"
)

const (
	ContentTypeSSE = "text/event-stream"
)

// Attempts to open an HTTP connection, validates that the server supports
// Server-Sent Events on the URL.
func httpConnectToSSE(url string) (*http.Response, error) {
	response, err := http.Get(url)
	if err != nil {
		return response, err
	}
	if response.StatusCode != http.StatusOK {
		return response, &ErrHttpNotOk{response.StatusCode}
	}
	contentType := response.Header.Get("Content-Type")
	if contentType != ContentTypeSSE {
		return response, &ErrContentType{contentType}
	}
	return response, nil
}
