package sse

import (
	"io"
)

const (
	STATUS_CONNECTING byte = iota
	STATUS_OPEN
	STATUS_CLOSED
)

type (
	EventSource interface {
		URL() (url string)
		ReadyState() (state byte)
		LastEventId() (id string)
		Events() (events <-chan Event)
		Close()
	}
	eventSource struct {
		lastEventId string
		url         string
		in          io.ReadCloser
		out         <-chan Event
		readyState  byte
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
	me.out = nil
	me.lastEventId = ""
	me.readyState = STATUS_CONNECTING
}

// Attempts to connect and updates internal status depending on the outcome.
func (me *eventSource) connect() (err error) {
	response, err := httpConnectToSSE(me.url)
	if err != nil {
		me.readyState = STATUS_CLOSED
		return err
	}
	me.in = response.Body
	me.consume()
	me.readyState = STATUS_OPEN
	return nil
}

// Method consume() must be called once connect() succeeds.
// It parses the input reader and assigns the event output channel accordingly.
func (me *eventSource) consume() {
	me.out = me.wrap(DefaultDecoder.Decode(me.in))
}

// Wraps an input of events, updates internal state for lastEventId
// and forwards the events to the final output.
func (me *eventSource) wrap(in <-chan Event) <-chan Event {
	out := make(chan Event)
	go func() {
		for {
			select {
			case ev, ok := <-in:
				if !ok {
					close(out)
					return
				}
				me.lastEventId = ev.Id()
				out <- ev
			}
		}
	}()
	return out
}

// Returns the event source URL.
func (me *eventSource) URL() string {
	return me.url
}

// Returns the event source connection state, either connecting, open or closed.
func (me *eventSource) ReadyState() byte {
	return me.readyState
}

// Returns the last event source Event id.
func (me *eventSource) LastEventId() string {
	return me.lastEventId
}

// Returns the channel of events. Events will be queued in the channel as they
// are received.
func (me *eventSource) Events() <-chan Event {
	return me.out
}

// Closes the event source.
// After closing the event source, it cannot be reused again.
func (me *eventSource) Close() {
	me.in.Close()
	me.readyState = STATUS_CLOSED
}
