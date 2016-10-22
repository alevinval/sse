package sse_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/mubit/sse"
	"github.com/mubit/sse/tests"
	"github.com/stretchr/testify/assert"
)

// Extracts decoder from a string
func newDecoder(data string) sse.Decoder {
	reader := bytes.NewReader([]byte(data))
	return sse.NewDecoder(reader)
}

func TestEOFIsReturned(t *testing.T) {
	decoder := newDecoder("")
	ev, err := decoder.Decode()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, ev)
}

func TestBigEventGrowsTheBuffer(t *testing.T) {
	bigEvent := tests.NewEventWithPadding(32000)
	decoder := newDecoder(string(bigEvent))

	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		actualLength := len(ev.Data()) + len(ev.Name()) + 8
		assert.Equal(t, 32000, actualLength)
	}
}

func TestEventNameAndData(t *testing.T) {
	decoder := newDecoder("event: some event\r\ndata: some event value\r\n\n")

	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "some event", ev.Name())
		assert.Equal(t, "some event value", string(ev.Data()))
	}
}

func TestEventNameAndDataManyEvents(t *testing.T) {
	decoder := newDecoder("event: first event\r\ndata: first value\r\n\nevent: second event\r\ndata: second value\r\n\n")

	ev1, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.NotNil(t, ev1)
		assert.Equal(t, "first event", ev1.Name())
		assert.Equal(t, "first value", string(ev1.Data()))
	}

	ev2, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.NotNil(t, ev2)
		assert.Equal(t, "second event", ev2.Name())
		assert.Equal(t, "second value", string(ev2.Data()))
	}
}

func TestStocksExample(t *testing.T) {
	decoder := newDecoder("data: YHOO\ndata: +2\ndata: 10\n\n")
	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "", ev.ID())
		assert.Equal(t, "YHOO\n+2\n10", string(ev.Data()))
	}
}

func TestFirstWhitespaceIsIgnored(t *testing.T) {
	decoder := newDecoder("data: first\n\ndata: second\n\n")

	ev1, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "first", string(ev1.Data()))
	}

	ev2, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "second", string(ev2.Data()))
	}
}

func TestOnlyOneWhitespaceIsIgnored(t *testing.T) {
	decoder := newDecoder("data:   first\n\n") // 3 whitespaces
	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "  first", string(ev.Data())) // 2 whitespaces
	}
}

func TestEventsWithNoDataThenWithNewLine(t *testing.T) {
	decoder := newDecoder("data\n\ndata\ndata\n\ndata:")

	ev1, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "", string(ev1.Data()))
	}

	ev2, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "\n", string(ev2.Data()))
	}
}

func TestCommentIsIgnoredAndDataIsNot(t *testing.T) {
	decoder := newDecoder(": test stream\n\ndata: first event\nid: 1\n\ndata:second event\nid\n\ndata:  third event\n\n")

	ev1, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "1", ev1.ID())
		assert.Equal(t, "first event", string(ev1.Data()))
	}

	ev2, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "", ev2.ID())
		assert.Equal(t, "second event", string(ev2.Data()))
	}

	ev3, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "", ev3.ID())
		assert.Equal(t, " third event", string(ev3.Data()))
	}
}

func TestOneLineDataParseWithDoubleRN(t *testing.T) {
	decoder := newDecoder("data: this is a test\r\n\r\n")

	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "this is a test", string(ev.Data()))
	}
}

func TestOneLineDataParseWithoutDoubleRN(t *testing.T) {
	decoder := newDecoder("data: this is a test\r\n\n")

	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "this is a test", string(ev.Data()))
	}
}

func TestTwoLinesDataParseWithRNAndDoubleRN(t *testing.T) {
	decoder := newDecoder("data: this is \r\ndata: a test\r\n\r\n")

	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "this is \na test", string(ev.Data()))
	}
}

func TestNewLineWithCR(t *testing.T) {
	decoder := newDecoder("event: name\ndata: some\rdata:  data\r\n\n")

	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "name", ev.Name())
		assert.Equal(t, "some\n data", string(ev.Data()))
	}
}

// Bug #4: decoder: pure CR not recognized as end of line #4
func TestPureLineFeedsWithCarriageReturn(t *testing.T) {
	decoder := newDecoder("event: name\rdata: some\rdata:  data\r\r")
	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "name", ev.Name())
		assert.Equal(t, "some\n data", string(ev.Data()))
	}
}
