package sse

import (
	"errors"
	"io"
	"net/http"
	"sync"
	"time"
)

type ReadyState uint16

const (
	AllowedContentType = "text/event-stream"

	// Connecting while trying to establish connection with the stream.
	Connecting ReadyState = iota - 1
	// Open after connection is established with the server.
	Open
	// Closing after Close is invoked.
	Closing
	// Closed after the connection is closed.
	Closed
)

var (
	ErrContentType = errors.New("eventsource: the content type of the stream is not allowed")
)

type (
	// EventSource connects and processes events from an SSE stream.
	EventSource interface {
		URL() (url string)
		ReadyState() (state ReadyState)
		MessageEvents() (events <-chan *MessageEvent)
		Close()
	}
	eventSource struct {
		url          string
		d            Decoder
		resp         *http.Response
		out          chan *MessageEvent

		// Status of the event stream.
		readyState    ReadyState
		readyStateMux sync.RWMutex
	}
)

// NewEventSource constructs returns an EventSource that satisfies the HTML5 EventSource specification.
func NewEventSource(url string) (EventSource, error) {
	es := eventSource{
		d:   nil,
		url: url,
		out: make(chan *MessageEvent),
	}
	return &es, es.connect()
}

// connect does a connection attempt, if the operation fails, attempt reconnecting
// according to the spec.
func (es *eventSource) connect() (err error) {
	es.setReadyState(Connecting)
	err = es.connectOnce()
	if err != nil {
		es.Close()
	}
	return
}

// reconnect to the stream several until the operation succeeds or the conditions
// to retry no longer hold true.
func (es *eventSource) reconnect() (err error) {
	es.setReadyState(Connecting)
	for es.mustReconnect(err) {
		time.Sleep(time.Duration(es.d.Retry()) * time.Millisecond)
		err = es.connectOnce()
	}
	if err != nil {
		es.Close()
	}
	return
}

// Attempts to connect and updates internal status depending on the outcome.
func (es *eventSource) connectOnce() (err error) {
	es.resp, err = es.doHttpConnect()
	if err != nil {
		return
	}
	es.setReadyState(Open)
	es.d = NewDecoder(es.resp.Body)
	go es.consume()
	return
}

func (es *eventSource) doHttpConnect() (*http.Response, error) {
	// Prepare request
	req, err := http.NewRequest("GET", es.url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", AllowedContentType)
	req.Header.Set("Cache-Control", "no-store")

	// Check response
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp, err
	}
	if resp.Header.Get("Content-Type") != AllowedContentType {
		return resp, ErrContentType
	}
	return resp, nil
}

// Method consume() must be called once connect() succeeds.
// It parses the input reader and assigns the event output channel accordingly.
func (es *eventSource) consume() {
	for {
		ev, err := es.d.Decode()
		if err != nil {
			if es.mustReconnect(err) {
				err = es.reconnect()
			}
			es.Close()
			return
		}
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
func (es *eventSource) ReadyState() ReadyState {
	es.readyStateMux.RLock()
	defer es.readyStateMux.RUnlock()
	return es.readyState
}

func (es *eventSource) setReadyState(newState ReadyState) {
	es.readyStateMux.Lock()
	defer es.readyStateMux.Unlock()

	// Once the EventSource is closed, its ready state cannot change anymore.
	if es.readyState == Closed {
		return
	}
	es.readyState = newState
}

// Returns the channel of events. MessageEvents will be queued in the channel as they
// are received.
func (es *eventSource) MessageEvents() <-chan *MessageEvent {
	return es.out
}

// Closes the event source.
// After closing the event source, it cannot be reused again.
func (es *eventSource) Close() {
	if es.acquireClosingRight() {
		if es.resp != nil {
			es.resp.Body.Close()
		}
		close(es.out)
		es.setReadyState(Closed)
	}
}

// Acquires closing right by setting readyState to Closing if no one else
// is attempting to close the EventSource.
func (es *eventSource) acquireClosingRight() bool {
	es.readyStateMux.Lock()
	defer es.readyStateMux.Unlock()
	if es.readyState == Closed || es.readyState == Closing {
		return false
	}
	es.readyState = Closing
	return true
}
