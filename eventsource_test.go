package sse_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mubit/sse"
	"github.com/mubit/sse/tests"
	"github.com/stretchr/testify/assert"
)

const (
	eventStream = "text/event-stream"
	textPlain   = "text/plain; charset=utf-8"
)

type handler struct {
	// Content Type that will be served by the test server.
	ContentType string

	// Maximum number of requests that will be served by the handler.
	// When maximum is reached, the handler will immediately return StatusNoContent
	// to properly indicate there is nothing left to stream.
	MaxRequests int

	t           *testing.T
	lastEventID string
	events      chan string
	closer      chan struct{}
	closed      chan struct{}
}

func newServer(t *testing.T) (*httptest.Server, *handler) {
	handler := &handler{
		ContentType: eventStream,
		MaxRequests: 1,
		t:           t,
		events:      make(chan string),
		closer:      make(chan struct{}),
		closed:      make(chan struct{}, 1),
	}
	return httptest.NewServer(handler), handler
}

func (h *handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		h.closed <- struct{}{}
	}()
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Content-Type", h.ContentType)

	// Assert sse.EventSource follows the spec and provides the Last-Event-ID header.
	if !assert.Equal(h.t, h.lastEventID, req.Header.Get("Last-Event-ID"), "spec violation: eventsource reconnected without providing the last event id.") {
		rw.WriteHeader(http.StatusNoContent)
		return
	}

	if h.MaxRequests <= 0 {
		rw.WriteHeader(http.StatusNoContent)
		return
	}
	h.MaxRequests--
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

func (h *handler) SendAndClose(ev *sse.MessageEvent) {
	h.Send(ev)
	h.Close()
}

func (h *handler) Send(ev *sse.MessageEvent) {
	h.SendString(tests.MessageEventToString(ev))
	h.lastEventID = ev.LastEventID
}

func (h *handler) SendString(data string) {
	h.events <- data
}

func (h *handler) Close() {
	h.closer <- struct{}{}
	<-h.closed
}

func closeTestServer(s *httptest.Server, h *handler) {
	// The test finished and we are cleaning up: force the handler to return on any
	// pending request.
	go h.Close()

	// Shutdown the test server.
	s.Close()
}

func TestEventSourceStates(t *testing.T) {
	for _, test := range []struct {
		stateNumber   byte
		expectedState sse.ReadyState
	}{
		{0, sse.Connecting},
		{1, sse.Open},
		{2, sse.Closing},
		{3, sse.Closed},
	} {
		assert.Equal(t, test.expectedState, sse.ReadyState(test.stateNumber))
	}
}

func TestNewEventSourceWithInvalidContentType(t *testing.T) {
	s, handler := newServer(t)
	defer closeTestServer(s, handler)
	handler.ContentType = textPlain

	es, err := sse.NewEventSource(s.URL)
	assert.Equal(t, sse.Connecting, <-es.ReadyState())
	assert.Equal(t, sse.Closing, <-es.ReadyState())
	assert.Equal(t, sse.Closed, <-es.ReadyState())

	if assert.Error(t, err) {
		assert.Equal(t, sse.ErrContentType, err)
		assert.Equal(t, s.URL, es.URL())
		_, ok := <-es.MessageEvents()
		assert.False(t, ok)
	}
	es.Close()
}

func TestNewEventSourceWithRightContentType(t *testing.T) {
	s, handler := newServer(t)
	defer closeTestServer(s, handler)

	es, err := sse.NewEventSource(s.URL)
	if assert.NoError(t, err) {
		return
	}
	assert.Equal(t, sse.Connecting, <-es.ReadyState())
	assert.Equal(t, sse.Open, <-es.ReadyState())

	expectedEv := tests.NewMessageEvent("", "", 128)
	go handler.Send(expectedEv)
	ev, ok := <-es.MessageEvents()
	if assert.True(t, ok) {
		assert.Equal(t, expectedEv.Data, ev.Data)
	}

	es.Close()
	assert.Equal(t, sse.Closing, <-es.ReadyState())
	assert.Equal(t, sse.Closed, <-es.ReadyState())
}

func TestNewEventSourceSendingEvent(t *testing.T) {
	s, handler := newServer(t)
	defer closeTestServer(s, handler)

	es, err := sse.NewEventSource(s.URL)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, sse.Connecting, <-es.ReadyState())
	assert.Equal(t, sse.Open, <-es.ReadyState())
	expectedEvent := tests.NewMessageEvent("", "", 1024)
	go handler.SendAndClose(expectedEvent)
	ev, ok := <-es.MessageEvents()
	if assert.True(t, ok) {
		assert.Equal(t, expectedEvent.Data, ev.Data)
	}
}

func TestEventSourceLastEventID(t *testing.T) {
	s, handler := newServer(t)
	defer closeTestServer(s, handler)

	es, err := sse.NewEventSource(s.URL)
	if !assert.NoError(t, err) {
		return
	}
	go drain(es.ReadyState())

	expectedEv := tests.NewMessageEvent("123", "", 512)
	go handler.Send(expectedEv)
	ev, ok := <-es.MessageEvents()
	if assert.True(t, ok) {
		assert.Equal(t, expectedEv.LastEventID, ev.LastEventID)
		assert.Equal(t, expectedEv.Data, ev.Data)
	}

	go handler.Send(tests.NewMessageEvent("", "", 32))
	ev, ok = <-es.MessageEvents()
	if assert.True(t, ok) {
		assert.Equal(t, expectedEv.LastEventID, ev.LastEventID)
	}

	es.Close()
}

func TestEventSourceRetryIsRespected(t *testing.T) {
	s, handler := newServer(t)
	defer closeTestServer(s, handler)
	handler.MaxRequests = 3

	es, err := sse.NewEventSource(s.URL)
	if !assert.NoError(t, err) {
		return
	}
	go drain(es.ReadyState())

	// Big retry
	handler.SendString(tests.NewRetryEvent(100))
	handler.Close()
	go handler.Send(tests.NewMessageEvent("", "", 128))
	select {
	case _, ok := <-es.MessageEvents():
		assert.True(t, ok)
	case <-timeout(150 * time.Millisecond):
		assert.Fail(t, "event source did not reconnect within the allowed time.")
	}

	// Smaller retry
	handler.SendString(tests.NewRetryEvent(1))
	handler.Close()
	go handler.Send(tests.NewMessageEvent("", "", 128))
	select {
	case _, ok := <-es.MessageEvents():
		assert.True(t, ok)
	case <-timeout(10 * time.Millisecond):
		assert.Fail(t, "event source did not reconnect within the allowed time.")
	}
	es.Close()
}

func TestDropConnectionCannotReconnect(t *testing.T) {
	s, handler := newServer(t)
	defer closeTestServer(s, handler)

	es, err := sse.NewEventSource(s.URL)
	assert.NoError(t, err)
	assert.Equal(t, sse.Connecting, <-es.ReadyState())
	assert.Equal(t, sse.Open, <-es.ReadyState())

	handler.Close()

	assert.Equal(t, sse.Connecting, <-es.ReadyState())
	assert.Equal(t, sse.Open, <-es.ReadyState())
	assert.Equal(t, sse.Closing, <-es.ReadyState())
	assert.Equal(t, sse.Closed, <-es.ReadyState())

	_, ok := <-es.MessageEvents()
	assert.False(t, ok)

	es.Close()
}

func TestDropConnectionCanReconnect(t *testing.T) {
	s, handler := newServer(t)
	defer closeTestServer(s, handler)
	handler.MaxRequests = 2

	es, err := sse.NewEventSource(s.URL)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, sse.Connecting, <-es.ReadyState())
	assert.Equal(t, sse.Open, <-es.ReadyState())
	handler.Close()
	assert.Equal(t, sse.Connecting, <-es.ReadyState())
	assert.Equal(t, sse.Open, <-es.ReadyState())
	go func() {
		time.Sleep(25 * time.Millisecond)
		handler.Send(tests.NewMessageEvent("", "", 128))
	}()
	_, ok := <-es.MessageEvents()
	assert.True(t, ok)
	es.Close()
}

func TestLastEventIDHeaderOnReconnecting(t *testing.T) {
	s, handler := newServer(t)
	defer closeTestServer(s, handler)
	handler.MaxRequests = 2

	es, err := sse.NewEventSource(s.URL)
	if !assert.NoError(t, err) {
		return
	}
	go drain(es.ReadyState())

	handler.SendString(tests.NewRetryEvent(1))
	expectedEv := tests.NewMessageEvent("abc", "", 128)
	go handler.SendAndClose(expectedEv)
	_, ok := <-es.MessageEvents()
	assert.True(t, ok)

	go handler.Send(tests.NewMessageEvent("def", "", 128))
	_, ok = <-es.MessageEvents()
	assert.True(t, ok)
	es.Close()
}

func timeout(d time.Duration) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		time.Sleep(d)
		ch <- struct{}{}
	}()
	return ch
}

func drain(ch <-chan sse.ReadyState) {
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		}
	}
}
