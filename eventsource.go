package sse

import (
	"io"
)

const (
	// StatusConnecting is the status of the EventSource before it tries to establish connection with the server.
	StatusConnecting byte = iota
	// StatusOpen after it connects to the server.
	StatusOpen
	// StatusClosed after the connection is closed.
	StatusClosed
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
		lastEventID string
		url         string
		in          io.ReadCloser
		out         <-chan Event
		readyState  byte
	}
)

// NewEventSource constructs returns an EventSource that satisfies the HTML5 EventSource specification.
func NewEventSource(url string) (EventSource, error) {
	es := &eventSource{}
	es.initialise(url)
	err := es.connect()
	return es, err
}

func (es *eventSource) initialise(url string) {
	es.url = url
	es.in = nil
	es.out = nil
	es.lastEventID = ""
	es.readyState = StatusConnecting
}

// Attempts to connect and updates internal status depending on the outcome.
func (es *eventSource) connect() (err error) {
	response, err := httpConnectToSSE(es.url)
	if err != nil {
		es.readyState = StatusClosed
		return err
	}
	es.in = response.Body
	es.consume()
	es.readyState = StatusOpen
	return nil
}

// Method consume() must be called once connect() succeeds.
// It parses the input reader and assigns the event output channel accordingly.
func (es *eventSource) consume() {
	es.out = es.wrap(DefaultDecoder.Decode(es.in))
}

// Wraps an input of events, updates internal state for lastEventId
// and forwards the events to the final output.
func (es *eventSource) wrap(in <-chan Event) <-chan Event {
	out := make(chan Event)
	go func() {
		for {
			select {
			case ev, ok := <-in:
				if !ok {
					close(out)
					return
				}
				es.lastEventID = ev.ID()
				out <- ev
			}
		}
	}()
	return out
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
	es.in.Close()
	es.readyState = StatusClosed
}
