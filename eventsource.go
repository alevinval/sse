package sse

import (
	"errors"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	AllowedContentType = "text/event-stream"

	// StatusConnecting is the status of the EventSource before it tries to establish connection with the server.
	StatusConnecting byte = iota
	// StatusOpen after it connects to the server.
	StatusOpen
	// StatusClosed after the connection is closed.
	StatusClosed

	defaultRetry = time.Duration(1000)
)

var (
	ErrContentType = errors.New("eventsource: the content type of the stream is not allowed")

	// Map used by decoders to be able to change the retry time of the eventSource when
	// a retry event is received.
	globalDecoderMap = map[Decoder]*eventSource{}
	retryMux         = sync.RWMutex{}
)

type (
	// EventSource connects and processes events from an SSE stream.
	EventSource interface {
		URL() (url string)
		ReadyState() (state byte)
		LastEventID() (id string)
		Events() (events <-chan Event)
		Close()
	}
	eventSource struct {
		lastEventID  string
		url          string
		resp         *http.Response
		out          chan Event
		closeOutOnce chan bool

		// Reconnection waiting time in milliseconds
		retry time.Duration

		// Status of the event stream.
		readyState byte
	}
)

// NewEventSource constructs returns an EventSource that satisfies the HTML5 EventSource specification.
func NewEventSource(url string) (EventSource, error) {
	es := eventSource{
		url:          url,
		out:          make(chan Event),
		closeOutOnce: make(chan bool),
		retry:        defaultRetry,
	}
	go es.closeOnce()
	return &es, es.connect()
}

// connect does a connection attempt, if the operation fails, attempt reconnecting
// according to the spec.
func (es *eventSource) connect() (err error) {
	es.readyState = StatusConnecting

	// Attempt first connection.
	err = es.connectOnce()
	if err == nil {
		return
	}

	// If the first connect attempt fails, begin the reconnection process.
	for es.mustReconnect(err) {
		time.Sleep(es.retry)
		err = es.connectOnce()
	}
	if err != nil {
		es.Close()
	}
	return
}

// Attempts to connect and updates internal status depending on the outcome.
func (es *eventSource) connectOnce() error {
	resp, err := http.Get(es.url)
	if err != nil {
		es.resp = nil
		return err
	}
	if resp.Header.Get("Content-Type") != AllowedContentType {
		return ErrContentType
	}
	es.resp = resp
	es.readyState = StatusOpen
	go es.consume()
	return err
}

// Method consume() must be called once connect() succeeds.
// It parses the input reader and assigns the event output channel accordingly.
func (es *eventSource) consume() {
	d := NewDecoder(es.resp.Body)

	retryMux.Lock()
	globalDecoderMap[d] = es
	retryMux.Unlock()

	for {
		ev, err := d.Decode()
		if err != nil {
			if es.mustReconnect(err) {
				err = es.connect()
				return
			}
			es.Close()
			return
		}
		es.lastEventID = ev.ID()
		es.out <- ev
	}
}

// Clients will reconnect if the connection is closed;
// a client can be told to stop reconnecting using the HTTP 204 No Content response code.
func (es *eventSource) mustReconnect(err error) bool {
	switch err {
	case ErrContentType:
		return false
	case io.ErrUnexpectedEOF:
		return true
	}
	if es.resp != nil && es.resp.StatusCode == http.StatusNoContent {
		return false
	}
	return true
}

// Returns the event source URL.
func (es *eventSource) URL() string {
	return es.url
}

// Returns the event source connection state, either connecting, open or closed.
func (es *eventSource) ReadyState() byte {
	return es.readyState
}

// Returns the last event source Event id.
func (es *eventSource) LastEventID() string {
	return es.lastEventID
}

// Returns the channel of events. Events will be queued in the channel as they
// are received.
func (es *eventSource) Events() <-chan Event {
	return es.out
}

// Closes the event source.
// After closing the event source, it cannot be reused again.
func (es *eventSource) Close() {
	if es.readyState == StatusClosed {
		return
	}
	es.readyState = StatusClosed
	if es.resp != nil {
		es.resp.Body.Close()
	}
	es.sendClose()
}

func (es *eventSource) sendClose() {
	select {
	case es.closeOutOnce <- true:
	default:
	}
}

func (es *eventSource) closeOnce() {
	select {
	case <-es.closeOutOnce:
		close(es.out)
	}
}
