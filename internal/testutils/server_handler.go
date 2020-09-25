package testutils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const contentTypeEventStream = "text/event-stream"

// MessageEvent presents the payload being parsed from an EventSource.
type MessageEvent struct {
	LastEventID string
	Name        string
	Data        string
}

// TestServerHandler used to emulate an http server that follows
// the SSE spec
type TestServerHandler struct {
	// Server instance of the test HTTP server
	Server *httptest.Server

	// URL of the HTTP test server
	URL string

	// Content Type that will be served by the test server.
	ContentType string

	// MaxRequestsToProcess before closing the stream.
	MaxRequestsToProcess int

	t           *testing.T
	lastEventID string
	events      chan string
	closer      chan struct{}
}

func (h *TestServerHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Content-Type", h.ContentType)

	// Assert EventSource follows the spec and provides the Last-Event-ID header.
	if !assert.Equal(h.t, h.lastEventID, req.Header.Get("Last-Event-ID"), "spec violation: eventsource reconnected without providing the last event id.") {
		rw.WriteHeader(http.StatusNoContent)
		return
	}

	if h.MaxRequestsToProcess <= 0 {
		rw.WriteHeader(http.StatusNoContent)
		return
	}
	h.MaxRequestsToProcess--
	f, _ := rw.(http.Flusher)
	f.Flush()

	for {
		select {
		case <-h.closer:
			return
		case event, ok := <-h.events:
			if !ok {
				return
			}
			rw.Write([]byte(event))
			f.Flush()
		}
	}
}

func (h *TestServerHandler) SendAndClose(ev *MessageEvent) {
	h.Send(ev)
	h.CloseActiveRequest()
}

func (h *TestServerHandler) Send(ev *MessageEvent) {
	h.sendString(MessageEventToString(ev))
	h.lastEventID = ev.LastEventID
}

func (h *TestServerHandler) SendRetry(ev *RetryEvent) {
	h.sendString(RetryEventToString(ev))
}

// CloseActiveRequest cancels the current request being served
func (h *TestServerHandler) CloseActiveRequest() {
	h.closer <- struct{}{}
}

// Close cancels both the active request being served and the underlying
// test HTTP server
func (h *TestServerHandler) Close() {
	go h.CloseActiveRequest()
	h.Server.Close()
}

func (h *TestServerHandler) sendString(data string) {
	h.events <- data
}

func NewDefaultTestServerHandler(t *testing.T) *TestServerHandler {
	handler := &TestServerHandler{
		URL:                  "",
		ContentType:          contentTypeEventStream,
		MaxRequestsToProcess: 1,
		t:                    t,
		events:               make(chan string),
		closer:               make(chan struct{}),
	}
	handler.Server = httptest.NewServer(handler)
	handler.URL = handler.Server.URL
	return handler
}
