package eventsource

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"sync"
	"time"

	"github.com/go-rfc/sse/pkg/base"
	"github.com/go-rfc/sse/pkg/decoder"
)

const (
	allowedContentType = "text/event-stream"
)

var (
	// ErrContentType error indicates the content-type header is not accepted
	ErrContentType = errors.New("eventsource: the content type of the stream is not allowed")

	// ErrUnauthorized error indicates an authorization error of the connection request
	ErrUnauthorized = errors.New("eventsource: connection is unauthorized")
)

type (
	// EventSource connects and processes events from an HTTP server-sent events stream.
	EventSource struct {
		url              string
		requestModifiers []RequestModifier
		lastEventID      string
		d                *decoder.Decoder
		resp             *http.Response
		closed           bool
		closedMutex      *sync.RWMutex
		out              chan *base.MessageEvent
		readyState       chan Status
	}
)

// NewEventSource connects and returns an EventSource.
// Supports request modifiers in case you need to update HTTP headers
// of the underlying request.
func NewEventSource(url string, requestModifiers ...RequestModifier) (*EventSource, error) {
	es := &EventSource{
		d:           nil,
		url:         url,
		out:         make(chan *base.MessageEvent),
		readyState:  make(chan Status, 128),
		closedMutex: new(sync.RWMutex),
	}
	es.requestModifiers = append(es.requestModifiers, requestModifiers...)
	return es, es.connect()
}

// connect does a connection attempt, if the operation fails, attempt reconnecting
// according to the spec.
func (es *EventSource) connect() (err error) {
	err = es.connectOnce()
	if err != nil {
		es.Close(err)
	}
	return
}

// reconnect to the stream several until the operation succeeds or the conditions
// to retry no longer hold true.
func (es *EventSource) reconnect() (err error) {
	for es.mustReconnect(err) {
		time.Sleep(time.Duration(es.d.Retry()) * time.Millisecond)
		err = es.connectOnce()
	}
	if err != nil {
		es.Close(err)
	}
	return
}

// Attempts to connect and updates internal status depending on the outcome.
func (es *EventSource) connectOnce() (err error) {
	es.readyState <- Status{Connecting, nil}
	es.resp, err = es.doHTTPConnect()
	if err != nil {
		return
	}
	es.readyState <- Status{Open, nil}
	es.d = decoder.New(es.resp.Body)
	go es.consume()
	return
}

func (es *EventSource) doHTTPConnect() (*http.Response, error) {
	// Prepare request
	req, err := http.NewRequest("GET", es.url, nil)
	if err != nil {
		return nil, err
	}

	// apply request modifiers
	for _, requestModifier := range es.requestModifiers {
		requestModifier(req)
	}

	req.Header.Set("Accept", allowedContentType)
	req.Header.Set("Cache-Control", "no-store")
	if es.lastEventID != "" {
		req.Header.Set("Last-Event-ID", es.lastEventID)
	}

	// Check response
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode == 401 {
		return resp, ErrUnauthorized
	}

	mediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil || mediaType != allowedContentType {
		return resp, ErrContentType
	}
	return resp, nil
}

// Method consume() must be called once connect() succeeds.
// It parses the input reader and assigns the event output channel accordingly.
func (es *EventSource) consume() {
	for {
		ev, err := es.d.Decode()
		if err != nil {
			if es.mustReconnect(err) {
				es.reconnect()
			} else {
				es.Close(err)
			}
			return
		}
		es.lastEventID = ev.LastEventID
		es.out <- ev
	}
}

// Clients will reconnect if the connection is closed;
// a client can be told to stop reconnecting using the HTTP 204 No Content response code.
func (es *EventSource) mustReconnect(err error) bool {
	es.closedMutex.RLock()
	defer es.closedMutex.RUnlock()
	if es.closed {
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

// URL returns the event source URL.
func (es *EventSource) URL() string {
	return es.url
}

// MessageEvents returns a channel of received events.
func (es *EventSource) MessageEvents() <-chan *base.MessageEvent {
	return es.out
}

// ReadyState exposes a channel with updates on the ready state
// of the event source.
// It must be consumed together with MessageEvents.
func (es *EventSource) ReadyState() <-chan Status {
	return es.readyState
}

// Close the event source. Once closed, the event source cannot be re-used again.
func (es *EventSource) Close(err error) {
	es.closedMutex.Lock()
	defer es.closedMutex.Unlock()
	if es.closed {
		return
	}
	es.readyState <- Status{Closing, err}
	es.closed = true

	if es.resp != nil {
		es.resp.Body.Close()
	}

	close(es.out)
	es.readyState <- Status{Closed, err}
}
