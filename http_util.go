package sse

import (
	"fmt"
	"net/http"
)

const (
	contentTypeEventStream = "text/event-stream"
)

type (
	ErrHttpNotOk struct {
		statusCode int
	}
	ErrHttpContentType struct {
		contentType string
	}
)

func (me ErrHttpNotOk) Error() string {
	return fmt.Sprintf("request status code was %d instead of %d", me.statusCode, http.StatusOK)
}

func (me ErrHttpContentType) Error() string {
	return fmt.Sprintf("content type is %q instead of %q", me.contentType, contentTypeEventStream)
}

// Attempts to open an HTTP connection, validates that the server supports
// Server-Sent Events on the URL.
func httpConnectToSSE(url string) (*http.Response, error) {
	response, err := http.Get(url)
	if err != nil {
		return response, err
	}
	if response.StatusCode != http.StatusOK {
		return response, ErrHttpNotOk{response.StatusCode}
	}
	contentType := response.Header.Get("Content-Type")
	if contentType != contentTypeEventStream {
		return response, ErrHttpContentType{contentType}
	}
	return response, nil
}
