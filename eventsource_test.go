package sse

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
}

func newServer(t *testing.T) (*httptest.Server, *handler) {
	handler := &handler{
		ContentType: eventStream,
		MaxRequests: 1,
		t:           t,
		events:      make(chan string),
		closer:      make(chan struct{}),
	}
	return httptest.NewServer(handler), handler
}

func (h *handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Content-Type", h.ContentType)

	// Assert EventSource follows the spec and provides the Last-Event-ID header.
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

func (h *handler) SendAndClose(ev *MessageEvent) {
	h.Send(ev)
	h.Close()
}

func (h *handler) Send(ev *MessageEvent) {
	h.SendString(messageEventToString(ev))
	h.lastEventID = ev.LastEventID
}

func (h *handler) SendString(data string) {
	h.events <- data
}

func (h *handler) Close() {
	h.closer <- struct{}{}
}

// Asserts an EventSource has Closed readyState after calling Close on it.
func assertCloseClient(t *testing.T, es *EventSource) bool {
	es.Close()
	maxWaits := 10
	var waits int
	for es.ReadyState() == Closing && waits < maxWaits {
		time.Sleep(25 * time.Millisecond)
		waits++
	}
	return assert.Equal(t, Closed, es.ReadyState())
}

// Asserts an EventSource has Open readyState.
func assertIsOpen(t *testing.T, es *EventSource, err error) bool {
	return assert.Nil(t, err) && assert.Equal(t, Open, es.ReadyState())
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
		expectedState ReadyState
	}{
		{0, Connecting},
		{1, Open},
		{2, Closing},
		{3, Closed},
	} {
		assert.Equal(t, test.expectedState, ReadyState(test.stateNumber))
	}
}

func TestNewEventSourceWithInvalidContentType(t *testing.T) {
	s, handler := newServer(t)
	defer closeTestServer(s, handler)
	handler.ContentType = textPlain

	es, err := NewEventSource(s.URL)
	if assert.Error(t, err) {
		assert.Equal(t, ErrContentType, err)
		assert.Equal(t, s.URL, es.URL())
		assert.Equal(t, Closed, es.ReadyState())
		_, ok := <-es.MessageEvents()
		assert.False(t, ok)
	}
	assertCloseClient(t, es)
}

func TestNewEventSourceWithRightContentType(t *testing.T) {
	s, handler := newServer(t)
	defer closeTestServer(s, handler)

	es, err := NewEventSource(s.URL)
	if assertIsOpen(t, es, err) {
		expectedEv := newMessageEvent("", "", 128)
		go handler.SendAndClose(expectedEv)
		ev, ok := <-es.MessageEvents()
		if assert.True(t, ok) {
			assert.Equal(t, expectedEv.Data, ev.Data)
		}
	}
	assertCloseClient(t, es)
}

func TestNewEventSourceSendingEvent(t *testing.T) {
	s, handler := newServer(t)
	defer closeTestServer(s, handler)

	es, err := NewEventSource(s.URL)
	if assertIsOpen(t, es, err) {
		expectedEvent := newMessageEvent("", "", 1024)
		go handler.SendAndClose(expectedEvent)
		ev, ok := <-es.MessageEvents()
		if assert.True(t, ok) {
			assert.Equal(t, expectedEvent.Data, ev.Data)
		}
	}
	assertCloseClient(t, es)
}

func TestEventSourceLastEventID(t *testing.T) {
	s, handler := newServer(t)
	defer closeTestServer(s, handler)

	es, err := NewEventSource(s.URL)
	if assertIsOpen(t, es, err) {
		expectedEv := newMessageEvent("123", "", 512)
		go handler.Send(expectedEv)
		ev, ok := <-es.MessageEvents()
		if assert.True(t, ok) {
			assert.Equal(t, expectedEv.LastEventID, ev.LastEventID)
			assert.Equal(t, expectedEv.Data, ev.Data)
		}

		go handler.Send(newMessageEvent("", "", 32))
		ev, ok = <-es.MessageEvents()
		if assert.True(t, ok) {
			assert.Equal(t, expectedEv.LastEventID, ev.LastEventID)
		}
	}
	assertCloseClient(t, es)
}

func TestEventSourceRetryIsRespected(t *testing.T) {
	s, handler := newServer(t)
	defer closeTestServer(s, handler)
	handler.MaxRequests = 3

	es, err := NewEventSource(s.URL)
	if assertIsOpen(t, es, err) {
		// Big retry
		handler.SendString(newRetryEvent(100))
		handler.Close()
		go handler.Send(newMessageEvent("", "", 128))
		select {
		case _, ok := <-es.MessageEvents():
			assert.True(t, ok)
		case <-timeout(150 * time.Millisecond):
			assert.Fail(t, "event source did not reconnect within the allowed time.")
		}

		// Smaller retry
		handler.SendString(newRetryEvent(1))
		handler.Close()
		go handler.Send(newMessageEvent("", "", 128))
		select {
		case _, ok := <-es.MessageEvents():
			assert.True(t, ok)
		case <-timeout(10 * time.Millisecond):
			assert.Fail(t, "event source did not reconnect within the allowed time.")
		}
	}
	assertCloseClient(t, es)
}

func TestDropConnectionCannotReconnect(t *testing.T) {
	s, handler := newServer(t)
	defer closeTestServer(s, handler)

	es, err := NewEventSource(s.URL)
	if assertIsOpen(t, es, err) {
		handler.Close()
		_, ok := <-es.MessageEvents()
		if assert.False(t, ok) {
			assert.Equal(t, Closed, es.ReadyState())
		}
	}
	assertCloseClient(t, es)
}

func TestDropConnectionCanReconnect(t *testing.T) {
	s, handler := newServer(t)
	defer closeTestServer(s, handler)
	handler.MaxRequests = 2

	es, err := NewEventSource(s.URL)
	if assertIsOpen(t, es, err) {
		handler.Close()
		go func() {
			time.Sleep(25 * time.Millisecond)
			handler.Send(newMessageEvent("", "", 128))
		}()
		_, ok := <-es.MessageEvents()
		if assert.True(t, ok) {
			assert.Equal(t, Open, es.ReadyState())
		}
	}
	assertCloseClient(t, es)
}

func TestLastEventIDHeaderOnReconnecting(t *testing.T) {
	s, handler := newServer(t)
	defer closeTestServer(s, handler)
	handler.MaxRequests = 2

	es, err := NewEventSource(s.URL)
	if assertIsOpen(t, es, err) {
		handler.SendString(newRetryEvent(1))
		expectedEv := newMessageEvent("abc", "", 128)
		go handler.SendAndClose(expectedEv)
		_, ok := <-es.MessageEvents()
		assert.True(t, ok)

		go handler.Send(newMessageEvent("def", "", 128))
		_, ok = <-es.MessageEvents()
		assert.True(t, ok)
	}
	assertCloseClient(t, es)
}

func timeout(d time.Duration) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		time.Sleep(d)
		ch <- struct{}{}
	}()
	return ch
}
