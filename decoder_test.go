package sse_test

import (
	"bytes"
	"github.com/mubit/sse"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// Timeouts after the specified milliseconds
func timeout(ms time.Duration) <-chan bool {
	ch := make(chan bool, 1)
	go func() {
		time.Sleep(ms * time.Millisecond)
		ch <- true
	}()
	return ch
}

// Extracts events from a string
func decode(data string) <-chan sse.Event {
	reader := bytes.NewReader([]byte(data))
	return sse.DefaultDecoder.Decode(reader)
}

// Attempts to consume an event from the decoding stream. Fails
// on timeouts or closed channel.
func consume(t *testing.T, events <-chan sse.Event) sse.Event {
	select {
	case ev, ok := <-events:
		if !ok {
			assert.Fail(t, "no more events to dispatch")
		}
		return ev
	case <-timeout(1000):
		assert.Fail(t, "timeout reached before dispatching event")
		return nil
	}
}

func TestStocksExample(t *testing.T) {
	events := decode("data: YHOO\ndata: +2\ndata: 10\n\n")
	ev := consume(t, events)
	assert.Equal(t, "YHOO\n+2\n10", string(ev.Data()))
}

func TestIgnoredSpaceProducesTwoIdenticalEvents(t *testing.T) {
	events := decode("data:test\n\ndata: test\n\n")
	ev1 := consume(t, events)
	assert.Equal(t, "test", string(ev1.Data()))
	ev2 := consume(t, events)
	assert.Equal(t, "test", string(ev2.Data()))
}

func TestTwoEventsExample(t *testing.T) {
	events := decode("data\n\ndata\ndata\n\ndata:")
	ev1 := consume(t, events)
	assert.Equal(t, "", string(ev1.Data()))
	ev2 := consume(t, events)
	assert.Equal(t, "\n", string(ev2.Data()))
}

func TestStream(t *testing.T) {
	events := decode(": test stream\n\ndata: first event\nid: 1\n\ndata:second event\nid\n\ndata:  third event\n\n")
	ev1 := consume(t, events)
	assert.Equal(t, "first event", string(ev1.Data()))
	ev2 := consume(t, events)
	assert.Equal(t, "second event", string(ev2.Data()))
	ev3 := consume(t, events)
	assert.Equal(t, " third event", string(ev3.Data()))
}

func TestOneLineDataParseWithDoubleRN(t *testing.T) {
	events := decode("data: this is a test\r\n\r\n")
	ev := consume(t, events)
	assert.Equal(t, "this is a test", string(ev.Data()))
}

func TestOneLineDataParseWithoutDoubleRN(t *testing.T) {
	events := decode("data: this is a test\r\n\n")
	ev := consume(t, events)
	assert.Equal(t, "this is a test", string(ev.Data()))
}

func TestTwoLinesDataParseWithRNAndDoubleRN(t *testing.T) {
	events := decode("data: this is \r\ndata: a test\r\n\r\n")
	ev := consume(t, events)
	assert.Equal(t, "this is \na test", string(ev.Data()))
}
