package testutils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const contentTypeEventStream = "text/event-stream; charset=utf-8"

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

	// Server requires basic authorization if username is set
	BasicAuth struct {
		Username string
		Password string
	}

	t           *testing.T
	lastEventID string
	events      chan string
	closer      chan struct{}
}

func (h *TestServerHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	// verify basic auth
	if len(h.BasicAuth.Username) > 0 {
		username, password, ok := req.BasicAuth()
		if !ok {
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if h.BasicAuth.Username != username || h.BasicAuth.Password != password {
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

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

func (h *TestServerHandler) Send(data string) {
	h.SendWithID(data, "")
}

func (h *TestServerHandler) SendWithID(data, lastEventID string) {
	h.events <- data
	h.lastEventID = lastEventID
}

func (h *TestServerHandler) SendAndClose(data string) {
	h.SendAndCloseWithID(data, "")
}

func (h *TestServerHandler) SendAndCloseWithID(data string, lastEventID string) {
	h.SendWithID(data, lastEventID)
	h.CloseActiveRequest()
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
