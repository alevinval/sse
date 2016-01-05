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
	Event struct {
		Type string
		Data []byte
	}
	EventSource struct {
		url        string
		in         io.ReadCloser
		readyState byte
		events     chan Event
	}
)

// Constructs a new EventSource struct that satisfies the HTML5
// EventSource interface.
func NewEventSource(url string) (*EventSource, error) {
	es := &EventSource{}
	es.initialise(url)
	err := es.connect()
	return es, err
}

func (me *EventSource) initialise(url string) {
	me.url = url
	me.in = nil
	me.events = make(chan Event)
	me.readyState = CONNECTING
}

// Attempts to connect and updates internal status depending on the outcome.
func (me *EventSource) connect() error {
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
func (me *EventSource) consume() {
	me.events = parseStream(me.in)
}

// Returns the event source URL.
func (me *EventSource) URL() string {
	return me.url
}

// Returns the event source connection state, either connecting, open or closed.
func (me *EventSource) ReadyState() byte {
	return me.readyState
}

// Returns the channel of events. Events will be queued in the channel as they
// are received.
func (me *EventSource) Events() <-chan Event {
	return me.events
}
