package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alevinval/sse/internal/testutils"
	"github.com/alevinval/sse/pkg/base"
	"github.com/alevinval/sse/pkg/decoder"
	"github.com/alevinval/sse/pkg/encoder"
	"github.com/stretchr/testify/assert"
)

const contentTypeEventStream = "text/event-stream; charset=utf-8"

// MockHandler used to emulate an http server that follows
// the SSE spec
type MockHandler struct {

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

	// Connected notifies when the connection has been accepted and is ready
	// to send events
	Connected chan struct{}

	t           *testing.T
	server      *httptest.Server
	encoder     *encoder.Encoder
	flusher     http.Flusher
	lastEventID string
	closer      chan struct{}
}

func NewMockHandler(t *testing.T) *MockHandler {
	handler := &MockHandler{
		URL:                  "",
		ContentType:          contentTypeEventStream,
		MaxRequestsToProcess: 1,
		t:                    t,
		closer:               make(chan struct{}),
		Connected:            make(chan struct{}, 1),
	}
	handler.server = httptest.NewServer(handler)
	handler.URL = handler.server.URL
	return handler
}

func (h *MockHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.encoder = encoder.New(rw)
	h.flusher = rw.(http.Flusher)

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

	if !assert.Equal(h.t, h.lastEventID, req.Header.Get("Last-Event-ID"), "spec violation: eventsource reconnected without providing the last event id.") {
		rw.WriteHeader(http.StatusNoContent)
		return
	}

	if h.MaxRequestsToProcess <= 0 {
		rw.WriteHeader(http.StatusNoContent)
		return
	}

	h.MaxRequestsToProcess--
	h.flusher.Flush()
	h.Connected <- struct{}{}

	select {
	case <-h.closer:
	case <-time.After(1 * time.Second):
		// No test ever should take more than 1 second to run
		h.t.Log("auto-closing active request after 1s")
		h.t.FailNow()
	}
}

func (h *MockHandler) WriteEvent(event *base.MessageEvent) {
	h.t.Logf("writing event %v", event)
	if event.ID != "" || event.HasID {
		h.lastEventID = event.ID
	}

	h.encoder.WriteComment("sending test event")
	h.encoder.WriteEvent(event)
	h.flusher.Flush()
}

func (h *MockHandler) WriteRetry(delayInMillis int, getDecoder func() *decoder.Decoder) {
	h.t.Logf("writing retry %d", delayInMillis)
	h.encoder.WriteRetry(delayInMillis)
	h.flusher.Flush()

	testutils.ExpectCondition(h.t, func() bool {
		actual := getDecoder().Retry().Milliseconds()
		h.t.Logf("expecting retry to be %d (actual=%d)", delayInMillis, actual)
		return actual == int64(delayInMillis)
	})
}

// CloseActiveRequest cancels the current request being served
func (h *MockHandler) CloseActiveRequest(block bool) {
	h.t.Logf("[closing active request (block=%v)]", block)
	if block {
		h.closer <- struct{}{}
	} else {
		select {
		case h.closer <- struct{}{}:
		default:
		}
	}
}

// Close cancels both the active request being served and the underlying
// test HTTP server
func (h *MockHandler) Close() {
	h.CloseActiveRequest(false)
	h.server.Close()
}
