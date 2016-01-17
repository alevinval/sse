package sse

import (
	"fmt"
	"net/http"
)

const (
	contentTypeEventStream = "text/event-stream"
)

type (
	// ErrHTTPNotOk is returned when a request returns a status code different than 200 OK.
	ErrHTTPNotOk struct {
		statusCode int
	}
	// ErrHTTPContentType is returned when a server does not accept the text/event-stream content type.
	ErrHTTPContentType struct {
		contentType string
	}
)

func (e ErrHTTPNotOk) Error() string {
	return fmt.Sprintf("request status code was %d instead of %d", e.statusCode, http.StatusOK)
}

func (e ErrHTTPContentType) Error() string {
	return fmt.Sprintf("content type is %q instead of %q", e.contentType, contentTypeEventStream)
}

// Attempts to open an HTTP connection, validates that the server supports
// Server-Sent Events on the URL.
func httpConnectToSSE(url string) (response *http.Response, err error) {
	response, err = http.Get(url)
	if err != nil {
		return response, err
	}
	if response.StatusCode != http.StatusOK {
		return response, ErrHTTPNotOk{response.StatusCode}
	}
	contentType := response.Header.Get("Content-Type")
	if contentType != contentTypeEventStream {
		return response, ErrHTTPContentType{contentType}
	}
	return response, nil
}
