package sse

import (
	"errors"
	"io"
	"net/http"
	"sync"
	"time"
)

type ReadyState byte

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
		ReadyState() (state ReadyState)
		LastEventID() (id string)
		Events() (events <-chan Event)
		Close()
	}
	eventSource struct {
		url          string
		d            Decoder
		resp         *http.Response
		out          chan Event
		closeOutOnce chan bool

		// Last recorded event ID
		lastEventID    string
		lastEventIDMux sync.RWMutex

		// Status of the event stream.
		readyState    ReadyState
		readyStateMux sync.RWMutex

		// Reconnection waiting time in milliseconds
		retry time.Duration
	}
)

// NewEventSource constructs returns an EventSource that satisfies the HTML5 EventSource specification.
func NewEventSource(url string) (EventSource, error) {
	es := eventSource{
		d:            nil,
		url:          url,
		out:          make(chan Event),
		closeOutOnce: make(chan bool),
		retry:        defaultRetry,
	}

	// Ensure the output channel is closed only once.
	go es.closeOnce()

	return &es, es.connect()
}

// connect does a connection attempt, if the operation fails, attempt reconnecting
// according to the spec.
func (es *eventSource) connect() (err error) {
	es.setReadyState(Connecting)
	err = es.connectOnce()
	if err != nil {
		err = es.reconnect()
	}
	return
}

// reconnect to the stream several until the operation succeeds or the conditions
// to retry no longer hold true.
func (es *eventSource) reconnect() (err error) {
	es.setReadyState(Connecting)
	for es.mustReconnect(err) {
		time.Sleep(es.retry * time.Millisecond)
		err = es.connectOnce()
	}
	if err != nil {
		es.Close()
	}
	return
}

// Attempts to connect and updates internal status depending on the outcome.
func (es *eventSource) connectOnce() (err error) {
	es.resp, err = es.httpConnect()
	if err != nil {
		return
	}
	delEventSource(es.d)
	es.d = NewDecoder(es.resp.Body)
	setEventSource(es.d, es)
	es.setReadyState(Open)
	go es.consume()
	return
}

func (es *eventSource) httpConnect() (*http.Response, error) {
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
		es.resp = nil
		return nil, err
	}
	if resp.Header.Get("Content-Type") != AllowedContentType {
		return nil, ErrContentType
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
				return
			}
			es.Close()
			return
		}
		id := ev.ID()
		if id != "" {
			es.setLastEventID(id)
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
	es.readyState = newState
}

// Returns the last event source Event id.
func (es *eventSource) LastEventID() string {
	es.lastEventIDMux.RLock()
	defer es.lastEventIDMux.RUnlock()
	return es.lastEventID
}

func (es *eventSource) setLastEventID(id string) {
	es.lastEventIDMux.Lock()
	defer es.lastEventIDMux.Unlock()
	es.lastEventID = id
}

// Returns the channel of events. Events will be queued in the channel as they
// are received.
func (es *eventSource) Events() <-chan Event {
	return es.out
}

// Closes the event source.
// After closing the event source, it cannot be reused again.
func (es *eventSource) Close() {
	state := es.ReadyState()
	if state == Closed || state == Closing {
		return
	}
	es.setReadyState(Closing)
	if es.resp != nil {
		es.resp.Body.Close()
	}
	es.sendClose()
	delEventSource(es.d)
	es.setReadyState(Closed)
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

func setEventSource(d Decoder, es *eventSource) {
	retryMux.Lock()
	defer retryMux.Unlock()
	globalDecoderMap[d] = es
}

func getEventSource(d Decoder) (es *eventSource, ok bool) {
	retryMux.RLock()
	defer retryMux.RUnlock()
	es, ok = globalDecoderMap[d]
	return
}

func delEventSource(d Decoder) {
	retryMux.Lock()
	defer retryMux.Unlock()
	delete(globalDecoderMap, d)
}
