package sse

import (
	"testing"
	"time"

	"github.com/go-rfc/sse/internal/testutils"
	"github.com/stretchr/testify/assert"
)

const (
	contentTypeTextPlain = "text/plain; charset=utf-8"
)

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

		es.Close()
		assert.Equal(t, Closed, es.ReadyState())
	})
}

func TestEventSourceConnectAndCloseThenReceive(t *testing.T) {
	runTest(t, func(handler *testutils.TestServerHandler) {
		url := handler.URL
		es, err := NewEventSource(url)

		assert.Nil(t, err)
		es.Close()

		_, ok := <-es.MessageEvents()
		assert.False(t, ok)
	})
}

func TestEventSourceWithInvalidContentType(t *testing.T) {
	runTest(t, func(handler *testutils.TestServerHandler) {
		handler.ContentType = contentTypeTextPlain
		es, err := NewEventSource(handler.URL)

		assert.Equal(t, ErrContentType, err)
		assert.Equal(t, Closed, es.ReadyState())
	})
}

func TestEventSourceConnectWriteAndReceiveShortEvent(t *testing.T) {
	runTest(t, func(handler *testutils.TestServerHandler) {
		es, err := NewEventSource(handler.URL)
		assert.Nil(t, err)

		expectedEv := testutils.NewMessageEvent("", "", 128)
		go handler.SendAndClose(expectedEv)

		ev, ok := <-es.MessageEvents()
		assert.True(t, ok)
		assert.Equal(t, expectedEv.Data, ev.Data)
	})
}

func TestEventSourceConnectWriteAndReceiveLongEvent(t *testing.T) {
	runTest(t, func(handler *testutils.TestServerHandler) {
		es, err := NewEventSource(handler.URL)
		assert.Nil(t, err)

		expectedEv := testutils.NewMessageEvent("", "", 128)
		go handler.SendAndClose(expectedEv)

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
		expected := testutils.NewMessageEvent(lastEventID, "", 512)
		go handler.Send(expected)

		actual, ok := <-es.MessageEvents()
		assert.True(t, ok)
		assert.Equal(t, lastEventID, actual.LastEventID)
		assert.Equal(t, expected.Data, actual.Data)

		go handler.Send(testutils.NewMessageEvent("", "", 32))

		actual, ok = <-es.MessageEvents()
		assert.Equal(t, lastEventID, actual.LastEventID)
	})
}

func TestEventSourceRetryIsRespected(t *testing.T) {
	runTest(t, func(handler *testutils.TestServerHandler) {
		handler.MaxRequestsToProcess = 3

		es, err := NewEventSource(handler.URL)
		assert.Nil(t, err)

		handler.SendRetry(testutils.NewRetryEvent(100))
		handler.CloseActiveRequest()
		go handler.Send(testutils.NewMessageEvent("", "", 128))
		select {
		case _, ok := <-es.MessageEvents():
			assert.True(t, ok)
		case <-timeout(125 * time.Millisecond):
			assert.Fail(t, "event source did not reconnect within the allowed time.")
		}

		// Smaller retry
		handler.SendRetry(testutils.NewRetryEvent(1))
		handler.CloseActiveRequest()
		go handler.Send(testutils.NewMessageEvent("", "", 128))
		select {
		case _, ok := <-es.MessageEvents():
			assert.True(t, ok)
		case <-timeout(10 * time.Millisecond):
			assert.Fail(t, "event source did not reconnect within the allowed time.")
		}
	})
}

func TestDropConnectionCannotReconnect(t *testing.T) {
	runTest(t, func(handler *testutils.TestServerHandler) {
		es, err := NewEventSource(handler.URL)
		assert.Nil(t, err)

		handler.CloseActiveRequest()

		_, ok := <-es.MessageEvents()
		assert.False(t, ok)
	})
}

func TestDropConnectionCanReconnect(t *testing.T) {
	runTest(t, func(handler *testutils.TestServerHandler) {
		handler.MaxRequestsToProcess = 2
		es, err := NewEventSource(handler.URL)
		assert.Nil(t, err)

		handler.CloseActiveRequest()
		time.Sleep(25 * time.Millisecond)
		go handler.Send(testutils.NewMessageEvent("", "", 128))
		_, ok := <-es.MessageEvents()
		assert.True(t, ok)
	})
}

func TestLastEventIDHeaderOnReconnecting(t *testing.T) {
	runTest(t, func(handler *testutils.TestServerHandler) {
		handler.MaxRequestsToProcess = 2
		es, err := NewEventSource(handler.URL)
		assert.Nil(t, err)

		handler.SendRetry(testutils.NewRetryEvent(1))

		// After closing, we retry and can poll the second message
		go handler.SendAndClose(testutils.NewMessageEvent("first", "", 128))
		_, ok := <-es.MessageEvents()
		assert.True(t, ok)
		assert.Equal(t, "first", es.lastEventID)

		go handler.Send(testutils.NewMessageEvent("second", "", 128))
		_, ok = <-es.MessageEvents()
		assert.True(t, ok)
		assert.Equal(t, "second", es.lastEventID)

	})
}

func timeout(d time.Duration) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		time.Sleep(d)
		ch <- struct{}{}
	}()
	return ch
}

type testFn = func(*testutils.TestServerHandler)

func runTest(t *testing.T, fn testFn) {
	t.Log("setting up test")
	h := testutils.NewDefaultTestServerHandler(t)
	defer h.Close()
	fn(h)
	t.Logf("tearing down test")
}
