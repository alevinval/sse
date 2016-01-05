package sse

import (
	"io"
)

const (
	CONNECTING byte = iota
	OPEN
	CLOSED
)

type (
	EventSource interface {
		URL() string
		ReadyState() byte
		Events() <-chan *Event
	}
	eventSource struct {
		url        string
		in         io.ReadCloser
		out        chan *Event
		readyState byte
	}
)

// Constructs a new EventSource struct that satisfies the HTML5
// EventSource interface.
func NewEventSource(url string) (EventSource, error) {
	es := &eventSource{}
	es.initialise(url)
	err := es.connect()
	return es, err
}

func (me *eventSource) initialise(url string) {
	me.url = url
	me.in = nil
	me.out = make(chan *Event)
	me.readyState = CONNECTING
}

// Attempts to connect and updates internal status depending on the outcome.
func (me *eventSource) connect() error {
	response, err := httpConnectToSSE(me.url)
	if err != nil {
		me.readyState = CLOSED
		return err
	}
	me.in = response.Body
	me.consume()
	me.readyState = OPEN
	return nil
}

// Method consume() must be called once connect() succeeds.
// It parses the input reader and assigns the event output channel accordingly.
func (me *eventSource) consume() {
	me.out = Decode(me.in)
}

// Returns the event source URL.
func (me *eventSource) URL() string {
	return me.url
}

// Returns the event source connection state, either connecting, open or closed.
func (me *eventSource) ReadyState() byte {
	return me.readyState
}

// Returns the channel of events. Events will be queued in the channel as they
// are received.
func (me *eventSource) Events() <-chan *Event {
	return me.out
}
