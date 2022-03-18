package decoder

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/go-rfc/sse/internal/testutils"
	"github.com/go-rfc/sse/pkg/base"
	"github.com/go-rfc/sse/pkg/encoder"
	"github.com/stretchr/testify/assert"
)

func TestEOFIsReturned(t *testing.T) {
	decoder := newDecoder("")
	ev, err := decoder.Decode()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, ev)
}

func TestBigEventGrowsTheBuffer(t *testing.T) {
	expectedEv := testutils.NewMessageEvent("", "", 32000)
	decoder := newDecoder(getMessageEventAsString(expectedEv))

	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, expectedEv.LastEventID, ev.LastEventID)
		assert.Equal(t, expectedEv.Name, ev.Name)
		assert.Equal(t, expectedEv.Data, ev.Data)
	}
}

func TestEventNameAndData(t *testing.T) {
	decoder := newDecoder("event: some event\r\ndata: some event value\r\n\n")

	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "some event", ev.Name)
		assert.Equal(t, "some event value", ev.Data)
	}
}

func TestEventNameAndDataManyEvents(t *testing.T) {
	decoder := newDecoder("event: first event\r\ndata: first value\r\n\nevent: second event\r\ndata: second value\r\n\n")

	ev1, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.NotNil(t, ev1)
		assert.Equal(t, "first event", ev1.Name)
		assert.Equal(t, "first value", ev1.Data)
	}

	ev2, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.NotNil(t, ev2)
		assert.Equal(t, "second event", ev2.Name)
		assert.Equal(t, "second value", ev2.Data)
	}
}

func TestStocksExample(t *testing.T) {
	decoder := newDecoder("data: YHOO\ndata: +2\ndata: 10\n\n")
	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "", ev.LastEventID)
		assert.Equal(t, "YHOO\n+2\n10", ev.Data)
	}
}

func TestFirstWhitespaceIsIgnored(t *testing.T) {
	decoder := newDecoder("data: first\n\ndata: second\n\n")

	ev1, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "first", ev1.Data)
	}

	ev2, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "second", ev2.Data)
	}
}

func TestOnlyOneWhitespaceIsIgnored(t *testing.T) {
	decoder := newDecoder("data:   first\n\n") // 3 whitespaces
	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "  first", ev.Data) // 2 whitespaces
	}
}

func TestEventsWithNoDataThenWithNewLine(t *testing.T) {
	decoder := newDecoder("data\n\ndata\ndata\n\ndata:")

	ev1, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "", ev1.Data)
	}

	ev2, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "\n", ev2.Data)
	}
}

func TestDecode_LastEventId_returnsId(t *testing.T) {
	stream := fmt.Sprintf("data: test\nid: valid id\n\n")
	decoder := newDecoder(stream)

	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "valid id", ev.LastEventID)
	}
}

func TestDecode_LastEventId_ignoresNullCharater(t *testing.T) {
	stream := fmt.Sprintf("data: test\nid: invalid id \u0000\n\n")
	decoder := newDecoder(stream)

	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "", ev.LastEventID)
	}
}

func TestCommentIsIgnoredAndDataIsNot(t *testing.T) {
	decoder := newDecoder(": test stream\n\ndata: first event\nid: 1\n\ndata:second event\nid\n\ndata:  third event\n\n")

	ev1, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "1", ev1.LastEventID)
		assert.Equal(t, "first event", ev1.Data)
	}

	ev2, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "", ev2.LastEventID)
		assert.Equal(t, "second event", ev2.Data)
	}

	ev3, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "", ev3.LastEventID)
		assert.Equal(t, " third event", ev3.Data)
	}
}

func TestOneLineDataParseWithDoubleRN(t *testing.T) {
	decoder := newDecoder("data: this is a test\r\n\r\n")

	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "this is a test", ev.Data)
	}
}

func TestOneLineDataParseWithoutDoubleRN(t *testing.T) {
	decoder := newDecoder("data: this is a test\r\n\n")

	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "this is a test", ev.Data)
	}
}

func TestTwoLinesDataParseWithRNAndDoubleRN(t *testing.T) {
	decoder := newDecoder("data: this is \r\ndata: a test\r\n\r\n")

	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "this is \na test", ev.Data)
	}
}

func TestNewLineWithCR(t *testing.T) {
	decoder := newDecoder("event: name\ndata: some\rdata:  data\r\n\n")

	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "name", ev.Name)
		assert.Equal(t, "some\n data", ev.Data)
	}
}

// Bug #4: decoder: pure CR not recognized as end of line #4
func TestPureLineFeedsWithCarriageReturn(t *testing.T) {
	decoder := newDecoder("event: name\rdata: some\rdata:  data\r\r")
	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "name", ev.Name)
		assert.Equal(t, "some\n data", ev.Data)
	}
}

func TestDecodeRetry(t *testing.T) {
	decoder := New(bytes.NewReader([]byte("retry: 100\nretry: a\n")))
	_, err := decoder.Decode()
	assert.Equal(t, time.Duration(100)*time.Millisecond, decoder.Retry())
	assert.Equal(t, io.EOF, err)
}

func BenchmarkDecodeEmptyEvent(b *testing.B) {
	runDecodingBenchmark(b, "data: \n\n")
}

func BenchmarkDecodeEmptyEventWithIgnoredLine(b *testing.B) {
	runDecodingBenchmark(b, ":ignored line \n\ndata: \n\n")
}

func BenchmarkDecodeShortEvent(b *testing.B) {
	runDecodingBenchmark(b, "data: short event\n\n")
}

func BenchmarkDecode1kEvent(b *testing.B) {
	ev := testutils.NewMessageEvent("", "", 1000)
	runDecodingBenchmark(b, getMessageEventAsString(ev))
}

func BenchmarkDecode4kEvent(b *testing.B) {
	ev := testutils.NewMessageEvent("", "", 4000)
	runDecodingBenchmark(b, getMessageEventAsString(ev))
}

func BenchmarkDecode8kEvent(b *testing.B) {
	ev := testutils.NewMessageEvent("", "", 8000)
	runDecodingBenchmark(b, getMessageEventAsString(ev))
}

func BenchmarkDecode16kEvent(b *testing.B) {
	ev := testutils.NewMessageEvent("", "", 16000)
	runDecodingBenchmark(b, getMessageEventAsString(ev))
}

func newDecoder(data string) *Decoder {
	reader := bytes.NewReader([]byte(data))
	return New(reader)
}

func runDecodingBenchmark(b *testing.B, data string) {
	reader := bytes.NewReader([]byte(data))
	b.ResetTimer()
	decoder := New(reader)
	for i := 0; i < b.N; i++ {
		decoder.Decode()
		reader.Seek(0, 0)
	}
}

func getMessageEventAsString(ev *base.MessageEvent) string {
	out := new(bytes.Buffer)
	e := encoder.New(out)
	e.WriteEvent(ev)
	return out.String()
}
