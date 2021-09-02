package sse

import (
	"testing"
	"time"

	"github.com/go-rfc/sse/internal/testutils"
	"github.com/stretchr/testify/assert"
)

const (
	contentTypeTextPlain = "text/plain; charset=utf-8"
	basicAuthUsername    = "foo"
	basicAuthPassword    = "bar"
)

// currentState holds the most recent value of the event source
var currentState ReadyState

func TestEventSourceStates(t *testing.T) {
	for _, test := range []struct {
		stateNumber   byte
		expectedState ReadyState
	}{
		{0, Connecting},
		{1, Open},
		{2, Closing},
		{3, Closed},
	} {
		assert.Equal(t, test.expectedState, ReadyState(test.stateNumber))
	}
}

func TestEventSourceConnectAndClose(t *testing.T) {
	runTest(t, func(handler *testutils.TestServerHandler) {
		url := handler.URL
		es, err := NewEventSource(url)

		assert.Nil(t, err)
		assert.Equal(t, url, es.URL())

		es.Close(nil)
		assertStates(t, []ReadyState{Connecting, Open, Closing, Closed}, es.ReadyState())
	})
}

func TestEventSourceConnectAndCloseThenReceive(t *testing.T) {
	runTest(t, func(handler *testutils.TestServerHandler) {
		url := handler.URL
		es, err := NewEventSource(url)

		assert.Nil(t, err)
		es.Close(nil)

		_, ok := <-es.MessageEvents()
		assert.False(t, ok)
	})
}

func TestEventSourceWithInvalidContentType(t *testing.T) {
	runTest(t, func(handler *testutils.TestServerHandler) {
		handler.ContentType = contentTypeTextPlain
		es, err := NewEventSource(handler.URL)

		assert.Equal(t, ErrContentType, err)
		assertStates(t, []ReadyState{Connecting, Closing, Closed}, es.ReadyState())
	})
}

func TestEventSourceConnectWriteAndReceiveShortEvent(t *testing.T) {
	runTest(t, func(handler *testutils.TestServerHandler) {
		es, err := NewEventSource(handler.URL)
		assert.Nil(t, err)

		expectedEv := newMessageEvent("", "", 128)
		go handler.SendAndClose(messageEventToString(expectedEv))

		ev, ok := <-es.MessageEvents()
		assert.True(t, ok)
		assert.Equal(t, expectedEv.Data, ev.Data)
	})
}

func TestEventSourceConnectWriteAndReceiveLongEvent(t *testing.T) {
	runTest(t, func(handler *testutils.TestServerHandler) {
		es, err := NewEventSource(handler.URL)
		assert.Nil(t, err)

		expectedEv := newMessageEvent("", "", 128)
		go handler.SendAndClose(messageEventToString(expectedEv))

		ev, ok := <-es.MessageEvents()
		assert.True(t, ok)
		assert.Equal(t, expectedEv.Data, ev.Data)
	})
}

func TestEventSourceLastEventID(t *testing.T) {
	runTest(t, func(handler *testutils.TestServerHandler) {
		es, err := NewEventSource(handler.URL)
		assert.Nil(t, err)

		lastEventID := "123"
		expected := newMessageEvent(lastEventID, "", 512)
		go handler.SendWithID(messageEventToString(expected), expected.LastEventID)

		actual, ok := <-es.MessageEvents()
		assert.True(t, ok)
		assert.Equal(t, lastEventID, actual.LastEventID)
		assert.Equal(t, expected.Data, actual.Data)

		ev := newMessageEvent("", "", 32)
		go handler.SendWithID(messageEventToString(ev), ev.LastEventID)

		actual, ok = <-es.MessageEvents()
		assert.Equal(t, lastEventID, actual.LastEventID)
	})
}

func TestEventSourceRetryIsRespected(t *testing.T) {
	runTest(t, func(handler *testutils.TestServerHandler) {
		handler.MaxRequestsToProcess = 3

		es, err := NewEventSource(handler.URL)
		assert.Nil(t, err)

		handler.SendAndClose(retryEventToString(100))
		go handler.Send(newMessageEventString("", "", 128))
		select {
		case _, ok := <-es.MessageEvents():
			assert.True(t, ok)
		case <-time.After(125 * time.Millisecond):
			assert.Fail(t, "event source did not reconnect within the allowed time.")
		}

		// Smaller retry
		handler.SendAndClose(retryEventToString(1))
		go handler.Send(newMessageEventString("", "", 128))
		select {
		case _, ok := <-es.MessageEvents():
			assert.True(t, ok)
		case <-time.After(10 * time.Millisecond):
			assert.Fail(t, "event source did not reconnect within the allowed time.")
		}

		assertStates(
			t,
			[]ReadyState{Connecting, Open, Connecting, Open, Connecting, Open},
			es.ReadyState(),
		)
	})
}

func TestEventSourceDropConnectionCannotReconnect(t *testing.T) {
	runTest(t, func(handler *testutils.TestServerHandler) {
		es, err := NewEventSource(handler.URL)
		assert.Nil(t, err)

		handler.CloseActiveRequest()

		_, ok := <-es.MessageEvents()
		assert.False(t, ok)
		assertStates(
			t,
			[]ReadyState{Connecting, Open, Connecting, Open, Closing, Closed},
			es.ReadyState(),
		)

	})
}

func TestEventSourceDropConnectionCanReconnect(t *testing.T) {
	runTest(t, func(handler *testutils.TestServerHandler) {
		handler.MaxRequestsToProcess = 2
		es, err := NewEventSource(handler.URL)
		assert.Nil(t, err)

		handler.CloseActiveRequest()
		time.Sleep(25 * time.Millisecond)
		go handler.Send(newMessageEventString("", "", 128))
		_, ok := <-es.MessageEvents()
		assert.True(t, ok)
		assertStates(
			t,
			[]ReadyState{Connecting, Open, Connecting, Open},
			es.ReadyState(),
		)
	})
}

func TestEventSourceLastEventIDHeaderOnReconnecting(t *testing.T) {
	runTest(t, func(handler *testutils.TestServerHandler) {
		handler.MaxRequestsToProcess = 2
		es, err := NewEventSource(handler.URL)
		assert.Nil(t, err)

		handler.Send(retryEventToString(1))

		// After closing, we retry and can poll the second message
		go handler.SendAndCloseWithID(newMessageEventString("first", "", 128), "first")
		_, ok := <-es.MessageEvents()
		assert.True(t, ok)
		assert.Equal(t, "first", es.lastEventID)

		go handler.SendWithID(newMessageEventString("second", "", 128), "second")
		_, ok = <-es.MessageEvents()
		assert.True(t, ok)
		assert.Equal(t, "second", es.lastEventID)
	})
}

func TestEventSourceWithBasicAuth(t *testing.T) {
	runTest(t, func(handler *testutils.TestServerHandler) {
		handler.BasicAuth.Username = basicAuthUsername
		handler.BasicAuth.Password = basicAuthPassword

		url := handler.URL
		es, err := NewEventSource(url, WithBasicAuth("foo", "bar"))

		assert.Nil(t, err)
		es.Close(nil)

		_, ok := <-es.MessageEvents()
		assert.False(t, ok)
	})
}

func TestEventSourceWithBasicAuthInvalidPassword(t *testing.T) {
	runTest(t, func(handler *testutils.TestServerHandler) {
		handler.BasicAuth.Username = basicAuthUsername
		handler.BasicAuth.Password = basicAuthPassword

		url := handler.URL
		es, err := NewEventSource(url, WithBasicAuth("foo", ""))

		assert.Equal(t, ErrUnauthorized, err)
		es.Close(nil)

		_, ok := <-es.MessageEvents()
		assert.False(t, ok)
	})
}

func assertStates(t *testing.T, expected []ReadyState, states <-chan Status) {
	actual := collectStates(states)
	assert.Equal(t, expected, actual)
}

func collectStates(states <-chan Status) []ReadyState {
	list := []ReadyState{}
	poll := true
	for poll {
		select {
		case status := <-states:
			list = append(list, status.ReadyState)
		case <-time.After(250 * time.Millisecond):
			poll = false
		}
	}
	return list
}

type testFn = func(*testutils.TestServerHandler)

func runTest(t *testing.T, fn testFn) {
	t.Log("setting up test")
	h := testutils.NewDefaultTestServerHandler(t)
	defer h.Close()
	fn(h)
	t.Logf("tearing down test")
}
