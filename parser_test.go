package sse

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStocksExample(t *testing.T) {
	data := []byte("data: YHOO\ndata: +2\ndata: 10\n\n")
	reader := bytes.NewReader(data)
	events := parseStream(reader)
	ev := <-events
	assert.Equal(t, "YHOO\n+2\n10", string(ev.Data))
}

func TestIgnoredSpaceProducesTwoIdenticalEvents(t *testing.T) {
	data := []byte("data:test\n\ndata: test\n\n")
	reader := bytes.NewReader(data)
	events := parseStream(reader)
	ev1 := <-events
	assert.Equal(t, "test", string(ev1.Data))
	ev2 := <-events
	assert.Equal(t, "test", string(ev2.Data))
}

func TestTwoEventsExample(t *testing.T) {
	data := []byte("data \n\ndata \ndata \n\ndata:")
	reader := bytes.NewReader(data)
	events := parseStream(reader)
	ev1 := <-events
	assert.Equal(t, "", string(ev1.Data))
	ev2 := <-events
	assert.Equal(t, "\n", string(ev2.Data))
}

func TestStream(t *testing.T) {
	data := []byte(": test stream\n\ndata: first event\nid: 1\n\ndata:second event\nid\n\ndata:  third event\n\n")
	reader := bytes.NewReader(data)
	events := parseStream(reader)
	ev1 := <-events
	assert.Equal(t, "first event", string(ev1.Data))
	ev2 := <-events
	assert.Equal(t, "second event", string(ev2.Data))
	ev3 := <-events
	assert.Equal(t, " third event", string(ev3.Data))
}

func TestOneLineDataParseWithDoubleRN(t *testing.T) {
	data := []byte("data: this is a test\r\n\r\n")
	reader := bytes.NewReader(data)
	events := parseStream(reader)
	ev := <-events
	assert.Equal(t, "this is a test", string(ev.Data))
}

func TestOneLineDataParseWithoutDoubleRN(t *testing.T) {
	data := []byte("data: this is a test\r\n\n")
	reader := bytes.NewReader(data)
	events := parseStream(reader)
	ev := <-events
	assert.Equal(t, "this is a test", string(ev.Data))
}

func TestTwoLinesDataParseWithRNAndDoubleRN(t *testing.T) {
	data := []byte("data: this is \r\ndata: a test\r\n\r\n")
	reader := bytes.NewReader(data)
	events := parseStream(reader)
	ev := <-events
	assert.Equal(t, "this is \na test", string(ev.Data))
}
