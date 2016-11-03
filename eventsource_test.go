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

	events chan []byte
	closer chan struct{}
}

func newServer() (*httptest.Server, *handler) {
	handler := &handler{
		ContentType: eventStream,
		MaxRequests: 1,
		events:      make(chan []byte),
		closer:      make(chan struct{}, 1),
	}
	return httptest.NewServer(handler), handler
}

func (s *handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Content-Type", s.ContentType)
	if s.MaxRequests <= 0 {
		rw.WriteHeader(http.StatusNoContent)
		return
	}
	s.MaxRequests--
	f, _ := rw.(http.Flusher)
	f.Flush()

	for {
		select {
		case event, ok := <-s.events:
			if !ok {
				return
			}
			rw.Write(event)
			f.Flush()
		case <-s.closer:
			return
		}
	}
}

func (s *handler) SendAndClose(data []byte) {
	s.Send(data)
	s.Close()
}

func (s *handler) Send(data []byte) {
	s.events <- data
}

func (s *handler) Close() {
	s.closer <- struct{}{}
}

// Asserts an sse.EventSource has Closed readyState after calling Close on it.
func assertCloseClient(t *testing.T, es sse.EventSource) bool {
	es.Close()
	maxWaits := 10
	var waits int
	for es.ReadyState() == sse.Closing && waits < maxWaits {
		time.Sleep(25 * time.Millisecond)
		waits++
	}
	return assert.Equal(t, sse.Closed, es.ReadyState())
}

// Asserts an sse.EventSource has Open readyState.
func assertIsOpen(t *testing.T, es sse.EventSource, err error) bool {
	return assert.Nil(t, err) && assert.Equal(t, sse.Open, es.ReadyState())
}

func closeTestServer(s *httptest.Server, h *handler) {
	// The test finished and we are cleaning up: force the handler to return on any
	// pending request.
	h.Close()

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
	s, handler := newServer()
	defer closeTestServer(s, handler)
	handler.ContentType = textPlain

	es, err := sse.NewEventSource(s.URL)
	if assert.Error(t, err) {
		assert.Equal(t, sse.ErrContentType, err)
		assert.Equal(t, s.URL, es.URL())
		assert.Equal(t, sse.Closed, es.ReadyState())
		_, ok := <-es.Events()
		assert.False(t, ok)
	}
	assertCloseClient(t, es)
}

func TestNewEventSourceWithRightContentType(t *testing.T) {
	s, handler := newServer()
	defer closeTestServer(s, handler)

	es, err := sse.NewEventSource(s.URL)
	if assertIsOpen(t, es, err) {
		ev := tests.NewEventWithPadding(128)
		go handler.SendAndClose(ev)
		recv, ok := <-es.Events()
		if assert.True(t, ok) {
			assert.Equal(t, tests.GetPaddedEventData(ev), recv.Data)
		}
	}
	assertCloseClient(t, es)
}

func TestNewEventSourceSendingEvent(t *testing.T) {
	s, handler := newServer()
	defer closeTestServer(s, handler)

	es, err := sse.NewEventSource(s.URL)
	if assertIsOpen(t, es, err) {
		expectedEvent := tests.NewEventWithPadding(2 << 10)
		go handler.SendAndClose(expectedEvent)
		ev, ok := <-es.Events()
		if assert.True(t, ok) {
			assert.Equal(t, tests.GetPaddedEventData(expectedEvent), ev.Data)
		}
	}
	assertCloseClient(t, es)
}

func TestEventSourceLastEventID(t *testing.T) {
	s, handler := newServer()
	defer closeTestServer(s, handler)

	es, err := sse.NewEventSource(s.URL)
	if assertIsOpen(t, es, err) {
		eventBytes := tests.NewEventWithPadding(2 << 8)
		expectedData := tests.GetPaddedEventData(eventBytes)
		eventBytes = append([]byte("id: 123\n"), eventBytes...)
		expectedID := "123"

		go handler.Send(eventBytes)
		ev, ok := <-es.Events()
		if assert.True(t, ok) {
			assert.Equal(t, expectedID, es.LastEventID())
			assert.Equal(t, expectedData, ev.Data)
		}

		go handler.Send(tests.NewEventWithPadding(32))
		_, ok = <-es.Events()
		if assert.True(t, ok) {
			assert.Equal(t, expectedID, es.LastEventID())
		}
	}
	assertCloseClient(t, es)
}

func TestEventSourceRetryIsRespected(t *testing.T) {
	s, handler := newServer()
	defer closeTestServer(s, handler)
	handler.MaxRequests = 3

	es, err := sse.NewEventSource(s.URL)
	if assertIsOpen(t, es, err) {
		// Big retry
		handler.Send([]byte("retry: 100\n"))
		handler.Close()
		go handler.Send(tests.NewEventWithPadding(128))
		select {
		case _, ok := <-es.Events():
			assert.True(t, ok)
		case <-timeout(150 * time.Millisecond):
			assert.Fail(t, "event source did not reconnect within the allowed time.")
		}

		// Smaller retry
		handler.Send([]byte("retry: 1\n"))
		handler.Close()
		go handler.Send(tests.NewEventWithPadding(128))
		select {
		case _, ok := <-es.Events():
			assert.True(t, ok)
		case <-timeout(10 * time.Millisecond):
			assert.Fail(t, "event source did not reconnect within the allowed time.")
		}
	}
	assertCloseClient(t, es)
}

func TestDropConnectionCannotReconnect(t *testing.T) {
	s, handler := newServer()
	defer closeTestServer(s, handler)

	es, err := sse.NewEventSource(s.URL)
	if assertIsOpen(t, es, err) {
		handler.Close()
		go handler.Send(tests.NewEventWithPadding(128))
		_, ok := <-es.Events()
		if assert.False(t, ok) {
			assert.Equal(t, sse.Closed, es.ReadyState())
		}
	}
	assertCloseClient(t, es)
}

func TestDropConnectionCanReconnect(t *testing.T) {
	s, handler := newServer()
	defer closeTestServer(s, handler)
	handler.MaxRequests = 2

	es, err := sse.NewEventSource(s.URL)
	if assertIsOpen(t, es, err) {
		handler.Close()
		go handler.Send(tests.NewEventWithPadding(128))
		_, ok := <-es.Events()
		if assert.True(t, ok) {
			assert.Equal(t, sse.Open, es.ReadyState())
		}
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
