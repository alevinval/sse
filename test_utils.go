package sse

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const contentTypeEventStream = "text/event-stream"

// retryEvent is used to represent a connection retry event
type retryEvent struct {
	delayInMs int
}

// testServerHandler used to emulate an http server that follows
// the SSE spec
type testServerHandler struct {
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

func (h *testServerHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
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

func (h *testServerHandler) SendAndClose(ev *MessageEvent) {
	h.Send(ev)
	h.CloseActiveRequest()
}

func (h *testServerHandler) Send(ev *MessageEvent) {
	h.sendString(messageEventToString(ev))
	h.lastEventID = ev.LastEventID
}

func (h *testServerHandler) SendRetry(ev *retryEvent) {
	h.sendString(retryEventToString(ev))
}

// CloseActiveRequest cancels the current request being served
func (h *testServerHandler) CloseActiveRequest() {
	h.closer <- struct{}{}
}

// Close cancels both the active request being served and the underlying
// test HTTP server
func (h *testServerHandler) Close() {
	go h.CloseActiveRequest()
	h.Server.Close()
}

func (h *testServerHandler) sendString(data string) {
	h.events <- data
}

func newTestServerHandler(t *testing.T) *testServerHandler {
	handler := &testServerHandler{
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

func newMessageEvent(lastEventID, name string, dataSize int) *MessageEvent {
	data := make([]byte, dataSize)
	for i := range data {
		data[i] = 'e'
	}
	return &MessageEvent{LastEventID: lastEventID, Name: name, Data: string(data)}
}

func newRetryEvent(delayInMs int) *retryEvent {
	return &retryEvent{delayInMs}
}

func messageEventToString(ev *MessageEvent) string {
	msg := ""
	if ev.LastEventID != "" {
		msg = buildString("id: ", ev.LastEventID, "\n")
	}
	if ev.Name != "" {
		msg = buildString(msg, "event: ", ev.Name, "\n")
	}
	return buildString(msg, "data: ", ev.Data, "\n\n")
}

func retryEventToString(ev *retryEvent) string {
	return buildString("retry: ", fmt.Sprintf("%d", ev.delayInMs), "\n")
}

func buildString(fields ...string) string {
	data := []byte{}
	for _, field := range fields {
		data = append(data, field...)
	}
	return string(data)
}
