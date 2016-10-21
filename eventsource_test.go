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
	ContentTypeTextStream = "text/event-stream"
	ContentTypeDefault    = "text/plain; charset=utf-8"
)

type SSE struct {
	ContentType string
	EventBytes  []byte
	Hang        bool
}

func (s *SSE) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Connection", "keep-alive")
	if s.ContentType != "" {
		rw.Header().Set("Content-Type", s.ContentType)
	}

	rw.Write(s.EventBytes)
	f, flusher := rw.(http.Flusher)
	if !flusher {
		panic("response writer must be flusher.")
	}
	f.Flush()

	if s.Hang {
		time.Sleep(10 * time.Second)
	}
	return
}

func CreateMockHTTPServer(contentType string, event []byte, hang bool) *httptest.Server {
	defer func() {
		// Attempt to avoid random failure on travis.
		time.Sleep(50 * time.Millisecond)
	}()
	s := SSE{ContentType: contentType, EventBytes: event, Hang: hang}
	return httptest.NewServer(&s)
}

func TestNewEventSourceWithInvalidContentType(t *testing.T) {
	s := CreateMockHTTPServer(ContentTypeDefault, []byte{}, false)
	es, err := sse.NewEventSource(s.URL)
	assert.Equal(t, s.URL, es.URL())
	assert.Equal(t, sse.ErrHTTPContentType{ContentType: ContentTypeDefault}, err)
	assert.Equal(t, sse.StatusClosed, es.ReadyState())
	_, ok := <-es.Events()
	assert.False(t, ok)
	es.Close()
}

func TestNewEventSourceWithRightContentType(t *testing.T) {
	s := CreateMockHTTPServer(ContentTypeTextStream, []byte{}, false)
	es, err := sse.NewEventSource(s.URL)
	assert.Nil(t, err)
	assert.Equal(t, sse.StatusOpen, es.ReadyState())
	_, ok := <-es.Events()
	assert.False(t, ok)
	es.Close()
	assert.Equal(t, sse.StatusClosed, es.ReadyState())
}

func TestNewEventSourceSendingEvent(t *testing.T) {
	expectedEvent := tests.NewEventWithPadding(2 << 10)
	s := CreateMockHTTPServer(ContentTypeTextStream, expectedEvent, false)
	es, err := sse.NewEventSource(s.URL)
	assert.Nil(t, err)
	assert.Equal(t, sse.StatusOpen, es.ReadyState())
	ev, ok := <-es.Events()
	assert.True(t, ok)
	if ok {
		assert.Equal(t, tests.GetPaddedEventData(expectedEvent), ev.Data())
	}
	es.Close()
	assert.Equal(t, sse.StatusClosed, es.ReadyState())
}

func TestNewEventSourceServerDropsConnection(t *testing.T) {
	s := CreateMockHTTPServer(ContentTypeTextStream, []byte{}, true)
	go func() {
		time.Sleep(250 * time.Millisecond)
		s.CloseClientConnections()
	}()
	es, err := sse.NewEventSource(s.URL)
	assert.Nil(t, err)
	assert.Equal(t, sse.StatusOpen, es.ReadyState())
	_, ok := <-es.Events()
	assert.False(t, ok)
	assert.Equal(t, sse.StatusClosed, es.ReadyState())
}

func TestEventSourceLastEventID(t *testing.T) {
	data := append([]byte("id: 123\n"), tests.NewEventWithPadding(2<<8)...)
	s := CreateMockHTTPServer(ContentTypeTextStream, data, false)
	es, err := sse.NewEventSource(s.URL)
	assert.Nil(t, err)
	assert.Equal(t, "", es.LastEventID())
	<-es.Events()
	assert.Equal(t, "123", es.LastEventID())
}
