package sse

import (
	"io"
	"net/http"
)

type (
	Event struct {
		Type string
		Data []byte
	}
	EventSource struct {
		url string
	}
)

// Constructs a new EventSource struct that satisfies the HTML5
// EventSource interface.
func NewEventSource(url string) *EventSource {
	es := &EventSource{}
	es.initialise(url)
	return es
}

func (me *EventSource) initialise(url string) {
	me.url = url
}

func (me *EventSource) URL() string {
	return me.url
}

func (me *EventSource) connect() (io.Reader, error) {
	response, err := http.Get(me.url)
	if err != nil {
		return nil, err
	}
	return response.Body, nil
}
func (me *EventSource) Consumer() (<-chan Event, error) {
	reader, err := me.connect()
	if err != nil {
		return nil, err
	}
	return parseStream(reader), nil
}
