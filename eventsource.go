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
		ReadyState() <-chan ReadyState
		MessageEvents() (events <-chan *MessageEvent)
		Close()
	}
	eventSource struct {
		url               string
		lastEventID       string
		d                 Decoder
		resp              *http.Response
		out               chan *MessageEvent
		readyStateUpdates chan ReadyState

		closed    bool
		closedMux sync.RWMutex
	}
)

// NewEventSource constructs returns an EventSource that satisfies the HTML5 EventSource specification.
func NewEventSource(url string) (EventSource, error) {
	es := eventSource{
		d:                 nil,
		url:               url,
		out:               make(chan *MessageEvent),
		readyStateUpdates: make(chan ReadyState, 3),
	}
	return &es, es.connect()
}

// connect does a connection attempt, if the operation fails, attempt reconnecting
// according to the spec.
func (es *eventSource) connect() (err error) {
	es.readyStateUpdates <- Connecting
	err = es.connectOnce()
	if err != nil {
		es.Close()
	}
	return
}

// reconnect to the stream several until the operation succeeds or the conditions
// to retry no longer hold true.
func (es *eventSource) reconnect() (err error) {
	es.readyStateUpdates <- Connecting
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
	es.resp, err = es.doHTTPConnect()
	if err != nil {
		return
	}
	es.readyStateUpdates <- Open
	es.d = NewDecoder(es.resp.Body)
	go es.consume()
	return
}

func (es *eventSource) doHTTPConnect() (*http.Response, error) {
	// Prepare request
	req, err := http.NewRequest("GET", es.url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", AllowedContentType)
	req.Header.Set("Cache-Control", "no-store")
	if es.lastEventID != "" {
		req.Header.Set("Last-Event-ID", es.lastEventID)
	}

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
		es.lastEventID = ev.LastEventID
		es.out <- ev
	}
}

// Clients will reconnect if the connection is closed;
// a client can be told to stop reconnecting using the HTTP 204 No Content response code.
func (es *eventSource) mustReconnect(err error) bool {
	es.closedMux.RLock()
	defer es.closedMux.RUnlock()
	if es.closed == true {
		return false
	}
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
func (es *eventSource) ReadyState() <-chan ReadyState {
	return es.readyStateUpdates
}

// Returns the channel of events. MessageEvents will be queued in the channel as they
// are received.
func (es *eventSource) MessageEvents() <-chan *MessageEvent {
	return es.out
}

// Closes the event source.
// After closing the event source, it cannot be reused again.
func (es *eventSource) Close() {
	es.closedMux.Lock()
	defer es.closedMux.Unlock()
	if es.closed == true {
		return
	}
	es.closed = true
	es.readyStateUpdates <- Closing
	if es.resp != nil {
		es.resp.Body.Close()
	}
	es.readyStateUpdates <- Closed
	close(es.readyStateUpdates)
	close(es.out)
}
