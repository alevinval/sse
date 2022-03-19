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
	ContentType = "text/event-stream"
)

var (
	// ErrContentType means the content-type header of the server is not the
	// expected one for an event-source. EventSource always expects
	// `text/event-stream`. content-type.
	ErrContentType = errors.New("eventsource: the content type of the stream is not allowed")

	// ErrUnauthorized means the server responded with an authorization error
	// status code.
	ErrUnauthorized = errors.New("eventsource: connection is unauthorized")
)

// EventSource connects and processes events from an HTTP server-sent
// events stream.
type EventSource struct {
	d                *decoder.Decoder
	resp             *http.Response
	closedMu         *sync.RWMutex
	readyState       chan Status
	out              chan *base.MessageEvent
	url              string
	lastEventID      string
	requestModifiers []RequestModifier
	closed           bool
}

// New EventSource, it accepts requests modifiers which allow to modify the
// underlying HTTP request, see RequestModifier.
func New(url string, requestModifiers ...RequestModifier) (*EventSource, error) {
	es := &EventSource{
		d:          nil,
		url:        url,
		out:        make(chan *base.MessageEvent),
		readyState: make(chan Status, 128),
		closedMu:   new(sync.RWMutex),
	}
	es.requestModifiers = append(es.requestModifiers, requestModifiers...)
	return es, es.connect()
}

// URL of the EventSource.
func (es *EventSource) URL() string {
	return es.url
}

// MessageEvents returns a receive-only channel where events are going to be
// sent to.
func (es *EventSource) MessageEvents() <-chan *base.MessageEvent {
	return es.out
}

// ReadyState exposes a channel with updates on the ready state of
// the EventSource. It must be consumed together with MessageEvents.
func (es *EventSource) ReadyState() <-chan Status {
	return es.readyState
}

// Close the event source.
// Once it has been closed, the event source cannot be re-used again.
func (es *EventSource) Close(err error) {
	es.closedMu.Lock()
	defer es.closedMu.Unlock()

	if es.closed {
		return
	}
	es.closed = true

	if es.resp != nil {
		es.resp.Body.Close()
	}

	close(es.out)
	es.readyState <- Status{ReadyState: Closed, Err: err}
}

func (es *EventSource) connect() (err error) {
	err = es.connectOnce()
	if err != nil {
		es.Close(err)
	}
	return
}

func (es *EventSource) reconnect() (err error) {
	for es.mustReconnect(err) {
		time.Sleep(es.d.Retry())
		err = es.connectOnce()
	}
	if err != nil {
		es.Close(err)
	}
	return
}

func (es *EventSource) connectOnce() (err error) {
	es.readyState <- Status{ReadyState: Connecting, Err: nil}
	es.resp, err = es.doHTTPConnect()
	if err != nil {
		return
	}
	es.readyState <- Status{ReadyState: Open, Err: nil}
	es.d = decoder.New(es.resp.Body)
	go es.consume()
	return
}

func (es *EventSource) doHTTPConnect() (*http.Response, error) {
	req, err := http.NewRequest("GET", es.url, nil)
	if err != nil {
		return nil, err
	}

	for _, requestModifier := range es.requestModifiers {
		requestModifier(req)
	}

	req.Header.Set("Accept", ContentType)
	req.Header.Set("Cache-Control", "no-store")
	if es.lastEventID != "" {
		req.Header.Set("Last-Event-ID", es.lastEventID)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode == 401 {
		return resp, ErrUnauthorized
	}

	mediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil || mediaType != ContentType {
		return resp, ErrContentType
	}
	return resp, nil
}

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
		if ev.ID != "" {
			es.lastEventID = ev.ID
		}
		es.out <- ev
	}
}

func (es *EventSource) mustReconnect(err error) bool {
	es.closedMu.RLock()
	defer es.closedMu.RUnlock()

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
