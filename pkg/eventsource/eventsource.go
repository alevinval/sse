package eventsource

import (
	"errors"
	"log"
	"mime"
	"net/http"
	"sync"
	"sync/atomic"
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
	readyState       chan Status
	out              chan *base.MessageEvent
	url              string
	lastEventID      string
	requestModifiers []RequestModifier

	safe struct {
		sync.RWMutex
		resp *http.Response
	}

	close struct {
		sync.Once
		notify    chan struct{}
		completed chan struct{}
		closed    uint32
	}
}

// New EventSource, it accepts requests modifiers which allow to modify the
// underlying HTTP request, see RequestModifier.
func New(url string, requestModifiers ...RequestModifier) (*EventSource, error) {
	es := &EventSource{
		url:        url,
		out:        make(chan *base.MessageEvent),
		readyState: make(chan Status, 128),

		safe: struct {
			sync.RWMutex
			resp *http.Response
		}{},
		close: struct {
			sync.Once
			notify    chan struct{}
			completed chan struct{}
			closed    uint32
		}{
			notify:    make(chan struct{}, 1),
			completed: make(chan struct{}, 1),
		},
	}
	es.requestModifiers = append(es.requestModifiers, requestModifiers...)

	initialConn := make(chan error)
	go es.consumer(initialConn)
	return es, <-initialConn
}

// URL of the EventSource.
func (es *EventSource) URL() string {
	return es.url
}

// MessageEvents returns a receive-only channel where events are received.
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
func (es *EventSource) Close() {
	es.doClose(nil)
	es.close.notify <- struct{}{}
	<-es.close.completed
}

func (es *EventSource) doClose(err error) {
	resp := es.getResp()
	if resp != nil {
		resp.Body.Close()
	}

	es.close.Do(func() {
		atomic.AddUint32(&es.close.closed, 1)
		es.readyState <- Status{ReadyState: Closed, Err: err}
	})
}

func (es *EventSource) connect() (err error) {
	es.readyState <- Status{ReadyState: Connecting, Err: nil}
	resp, err := es.doHTTPConnect()
	if err != nil {
		return
	}
	es.readyState <- Status{ReadyState: Open, Err: nil}
	es.setResp(resp)
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

func (es *EventSource) consumer(initialConn chan error) {
	defer func() {
		close(es.out)
		es.close.completed <- struct{}{}
	}()

	err := es.connect()
	if err != nil {
		es.doClose(err)
		initialConn <- err
		return
	}
	initialConn <- nil

	d := decoder.New(es.getResp().Body)
	for {
		ev, err := d.Decode()
		if err != nil {
			for es.mustReconnect(err) {
				time.Sleep(d.Retry())
				err = es.connect()
			}
			if es.isClosed() {
				return
			} else if err != nil {
				es.doClose(err)
				return
			}
			d = decoder.New(es.getResp().Body)
			continue
		}

		if ev.HasID {
			es.lastEventID = ev.ID
		}

		var sent bool
		for !sent {
			select {
			case <-es.close.notify:
				return
			case es.out <- ev:
				sent = true
			case <-time.After(1 * time.Second):
				log.Printf("eventsource: slow consumer, messages are not being consumed")
			}
		}
	}
}

func (es *EventSource) mustReconnect(err error) bool {
	if es.isClosed() {
		return false
	}

	resp := es.getResp()
	if resp != nil && resp.StatusCode == http.StatusNoContent {
		return false
	}

	switch err {
	case nil:
		return false
	case ErrContentType:
		return false
	default:
		return true
	}
}

func (es *EventSource) setResp(resp *http.Response) {
	es.safe.Lock()
	defer es.safe.Unlock()
	es.safe.resp = resp
}

func (es *EventSource) getResp() *http.Response {
	es.safe.RLock()
	defer es.safe.RUnlock()
	return es.safe.resp
}

func (es *EventSource) isClosed() bool {
	return atomic.LoadUint32(&es.close.closed) > 0
}
