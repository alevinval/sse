package eventsource

import (
	"testing"
	"time"

	"github.com/go-rfc/sse/internal/testutils/server"
	"github.com/go-rfc/sse/pkg/base"
	"github.com/stretchr/testify/assert"
)

func TestEventSource_WhenConnect_ThenNoError(t *testing.T) {
	setUp(t, func(handler *server.MockHandler) {
		sut, err := New(handler.URL)
		defer sut.Close()

		assert.NoError(t, err)
	})
}

func TestEventSource_WhenURL_ThenMatches(t *testing.T) {
	setUp(t, func(handler *server.MockHandler) {
		sut, _ := New(handler.URL)
		defer sut.Close()

		assert.Equal(t, handler.URL, sut.URL())
	})
}

func TestEventSource_WhenConnectAndClose_ThenReadyStatesMatch(t *testing.T) {
	setUp(t, func(handler *server.MockHandler) {
		sut, _ := New(handler.URL)

		<-handler.Connected
		sut.Close()

		assertStates(t, []ReadyState{Connecting, Open, Closed}, sut)
	})
}

func TestEventSource_WhenClose_ThenChannelIsClosed(t *testing.T) {
	setUp(t, func(handler *server.MockHandler) {
		sut, _ := New(handler.URL)
		sut.Close()

		_, ok := <-sut.MessageEvents()
		assert.False(t, ok)
	})
}

func TestEventSource_WhenInvalidContentType_ThenReturnsError(t *testing.T) {
	setUp(t, func(handler *server.MockHandler) {
		handler.ContentType = "text/plain; charset=utf-8"
		sut, err := New(handler.URL)
		defer sut.Close()

		assert.Equal(t, ErrContentType, err)
		assertStates(t, []ReadyState{Connecting, Closed}, sut)
	})
}

func TestEventSource_WhenReceiveEvent_ThenEventIsReceived(t *testing.T) {
	expected := &base.MessageEvent{
		ID:   "event-id",
		Name: "event-name",
		Data: "event-data",
	}
	setUp(t, func(handler *server.MockHandler) {
		sut, _ := New(handler.URL)
		defer sut.Close()

		handler.WriteEvent(expected)

		assertReceive(t, sut, expected)
	})
}

func TestEventSource_WhenMultipleWrites_ThenKeepsLastEventID(t *testing.T) {
	eventWithID := &base.MessageEvent{
		ID: "event-id",
	}
	eventWithoutID := &base.MessageEvent{
		Data: "event-data",
	}
	setUp(t, func(handler *server.MockHandler) {
		handler.MaxRequestsToProcess = 2
		sut, _ := New(handler.URL)
		defer sut.Close()

		<-handler.Connected
		handler.WriteRetry(1)
		handler.WriteEvent(eventWithID)
		assertReceive(t, sut, eventWithID)
		assert.Equal(t, "event-id", sut.lastEventID)

		// Force reconnection, which should set Last-Event-ID header
		handler.CloseActiveRequest(true)
		<-handler.Connected

		handler.WriteEvent(eventWithoutID)
		assertReceive(t, sut, eventWithoutID)
		assert.Equal(t, "event-id", sut.lastEventID)
	})
}

func TestEventSource_WhenMultipleWrites_ThenResetsLastEventID(t *testing.T) {
	eventWithID := &base.MessageEvent{
		ID: "event-id",
	}
	eventWithEmptyID := &base.MessageEvent{
		Data:  "event-data",
		HasID: true,
	}
	setUp(t, func(handler *server.MockHandler) {
		sut, _ := New(handler.URL)
		defer sut.Close()

		<-handler.Connected
		handler.WriteEvent(eventWithID)
		assertReceive(t, sut, eventWithID)
		assert.Equal(t, "event-id", sut.lastEventID)

		handler.WriteEvent(eventWithEmptyID)
		assertReceive(t, sut, eventWithEmptyID)
		assert.Equal(t, "", sut.lastEventID)
	})
}

func TestEventSource_WhenReconnecting_RetryIsRespected(t *testing.T) {
	longDelay := 50 * time.Millisecond
	shortDelay := 1 * time.Millisecond
	setUp(t, func(handler *server.MockHandler) {
		handler.MaxRequestsToProcess = 3

		sut, _ := New(handler.URL)
		defer sut.Close()

		<-handler.Connected
		handler.WriteRetry(int(longDelay.Milliseconds()))
		handler.CloseActiveRequest(true)
		assertConnectionWithinDeadline(t, handler, longDelay, 2*longDelay)

		handler.WriteRetry(int(shortDelay.Milliseconds()))
		handler.CloseActiveRequest(true)
		assertConnectionWithinDeadline(t, handler, shortDelay, longDelay)

		assertStates(
			t,
			[]ReadyState{Connecting, Open, Connecting, Open, Connecting, Open},
			sut,
		)
	})
}

func TestEventSource_WhenConnectionDropped_CannotReconnect(t *testing.T) {
	setUp(t, func(handler *server.MockHandler) {
		sut, _ := New(handler.URL)
		defer sut.Close()

		<-handler.Connected
		handler.WriteRetry(1)
		handler.CloseActiveRequest(true)

		assertNoReceives(t, sut)
		assertStates(
			t,
			[]ReadyState{Connecting, Open, Connecting, Open, Closed},
			sut,
		)
	})
}

func TestEventSource_DropConnection_CanReconnect(t *testing.T) {
	setUp(t, func(handler *server.MockHandler) {
		handler.MaxRequestsToProcess = 2
		sut, _ := New(handler.URL)
		defer sut.Close()

		<-handler.Connected
		handler.WriteRetry(1)
		handler.CloseActiveRequest(true)
		<-handler.Connected

		assertStates(
			t,
			[]ReadyState{Connecting, Open, Connecting, Open},
			sut,
		)
	})
}

func TestEventSource_WithBasicAuth(t *testing.T) {
	setUp(t, func(handler *server.MockHandler) {
		handler.BasicAuth.Username = "foo"
		handler.BasicAuth.Password = "bar"

		sut, _ := New(handler.URL, WithBasicAuth("foo", "bar"))
		defer sut.Close()

		<-handler.Connected
	})
}

func TestEventSource_WithBasicAuth_InvalidPassword(t *testing.T) {
	setUp(t, func(handler *server.MockHandler) {
		handler.BasicAuth.Username = "foo"
		handler.BasicAuth.Password = "bar"

		sut, err := New(handler.URL, WithBasicAuth("foo", ""))
		defer sut.Close()

		assert.Equal(t, ErrUnauthorized, err)
	})
}

func setUp(t *testing.T, test func(*server.MockHandler)) {
	handler := server.NewMockHandler(t)
	defer handler.Close()

	test(handler)
}

func assertNoReceives(t *testing.T, sut *EventSource) {
	_, ok := <-sut.MessageEvents()
	assert.False(t, ok, "expected no received, instead a received happened")
}

func assertReceive(t *testing.T, sut *EventSource, expected *base.MessageEvent) {
	actual, ok := <-sut.MessageEvents()
	if assert.True(t, ok, "expected to receive an event") {
		assert.Equal(t, expected.ID, actual.ID, "expected event id to match")
		assert.Equal(t, expected.Name, actual.Name, "expected event name to match")
		assert.Equal(t, expected.Data, actual.Data, "expected event data to match")
	}
}

func assertConnectionWithinDeadline(t *testing.T, handler *server.MockHandler, minDeadline, maxDeadline time.Duration) {
	start := time.Now()
	select {
	case <-handler.Connected:
		elapsed := time.Since(start)
		if elapsed < minDeadline {
			assert.Failf(t, "event source reconnected earlier than expected", "elapsed: %s", elapsed.String())
		}
	case <-time.After(maxDeadline):
		elapsed := time.Since(start)
		assert.Fail(t, "event source did not reconnect within the allowed time.", "elapsed: %s", elapsed.String())
	}
}

func assertStates(t *testing.T, expected []ReadyState, es *EventSource) {
	actual := collectStates(es.ReadyState())
	assert.Equal(t, stateToString(expected), stateToString(actual))
}

func collectStates(states <-chan Status) []ReadyState {
	list := []ReadyState{}
	for {
		select {
		case status := <-states:
			list = append(list, status.ReadyState)
			break
		case <-time.After(10 * time.Millisecond):
			return list
		}
	}
}

func stateToString(states []ReadyState) []string {
	list := []string{}
	for _, state := range states {
		list = append(list, state.String())
	}
	return list
}
