package decoder

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/alevinval/sse/internal/testutils"
	"github.com/alevinval/sse/pkg/encoder"
	"github.com/stretchr/testify/assert"
)

func TestDecoder_EmptyInput_ReturnsEOF(t *testing.T) {
	sut := newDecoder("")
	_, err := sut.Decode()
	assert.Equal(t, io.EOF, err)
}

func TestDecoder_EventEmptyID(t *testing.T) {
	sut := newDecoder("id\r\n\n")

	actual, err := sut.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "", actual.ID)
		assert.True(t, actual.HasID)
	}
}

func TestDecoder_EventID(t *testing.T) {
	sut := newDecoder("id: event-id\r\n\n")

	actual, err := sut.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "event-id", actual.ID)
		assert.True(t, actual.HasID)
	}
}

func TestDecoder_EventID_IgnoresNullCharacter(t *testing.T) {
	sut := newDecoder("data: test\nid: invalid id \u0000\n\n")

	actual, err := sut.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "", actual.ID)
		assert.False(t, actual.HasID)
	}
}

func TestDecoder_EventName(t *testing.T) {
	sut := newDecoder("event: event-name\r\n\n")

	actual, err := sut.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "event-name", actual.Name)
		assert.False(t, actual.HasID)
	}
}

func TestDecoder_EventData(t *testing.T) {
	sut := newDecoder("id: event-id\nevent: event-name\r\ndata: event-data\r\n\n")

	actual, err := sut.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "event-id", actual.ID)
		assert.Equal(t, "event-name", actual.Name)
		assert.Equal(t, "event-data", actual.Data)
		assert.True(t, actual.HasID)
	}
}

func TestDecoder_MultipleEvents(t *testing.T) {
	sut := newDecoder("id\nevent: first event\r\ndata: first value\r\n\nevent: second event\r\ndata: second value\r\n\n")

	first, err := sut.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "first event", first.Name)
		assert.Equal(t, "first value", first.Data)
		assert.True(t, first.HasID)
	}

	second, err := sut.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "second event", second.Name)
		assert.Equal(t, "second value", second.Data)
		assert.False(t, second.HasID)
	}
}

func TestDecoder_StocksExample(t *testing.T) {
	sut := newDecoder("data: YHOO\ndata: +2\ndata: 10\n\n")
	actual, err := sut.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "", actual.ID)
		assert.Equal(t, "YHOO\n+2\n10", actual.Data)
	}
}

func TestDecoder_OnlyOneWhitespaceIsIgnored(t *testing.T) {
	sut := newDecoder("data:   event-data\n\n") // 3 whitespaces
	actual, err := sut.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "  event-data", actual.Data) // 2 whitespaces
	}
}

func TestDecoder_EventsWithNoDataThenWithNewLine(t *testing.T) {
	sut := newDecoder("data\n\ndata\ndata\n\ndata:")

	actual, err := sut.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "", actual.Data)
	}

	actual, err = sut.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "\n", actual.Data)
	}

	_, err = sut.Decode()
	assert.ErrorIs(t, io.EOF, err)
}
func TestDecode_CommentIsIgnoredAndDataIsNot(t *testing.T) {
	sut := newDecoder(": test stream\n\ndata: first event\nid: 1\n\ndata:second event\nid\n\ndata:  third event\n\n")

	actual, err := sut.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "1", actual.ID)
		assert.Equal(t, "first event", actual.Data)
	}

	actual, err = sut.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "", actual.ID)
		assert.Equal(t, "second event", actual.Data)
	}

	actual, err = sut.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "", actual.ID)
		assert.Equal(t, " third event", actual.Data)
	}
}

func TestDecoder_OneLineDataParseWithDoubleRN(t *testing.T) {
	sut := newDecoder("data: this is a test\r\n\r\n")

	actual, err := sut.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "this is a test", actual.Data)
	}
}

func TestDecoder_OneLineDataParseWithoutDoubleRN(t *testing.T) {
	decoder := newDecoder("data: this is a test\r\n\n")

	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "this is a test", ev.Data)
	}
}

func TestDecoder_TwoLinesDataParseWithRNAndDoubleRN(t *testing.T) {
	sut := newDecoder("data: this is \r\ndata: a test\r\n\r\n")

	actual, err := sut.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "this is \na test", actual.Data)
	}
}

func TestDecoder_NewLineWithCR(t *testing.T) {
	sut := newDecoder("event: name\ndata: some\rdata:  data\r\n\n")

	ev, err := sut.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "name", ev.Name)
		assert.Equal(t, "some\n data", ev.Data)
	}
}

// Bug #4: decoder: pure CR not recognized as end of line #4
func TestDecoder_PureLineFeedsWithCarriageReturn(t *testing.T) {
	sut := newDecoder("event: name\rdata: some\rdata:  data\r\r")

	actual, err := sut.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, "name", actual.Name)
		assert.Equal(t, "some\n data", actual.Data)
	}
}

func TestDecoder_Retry(t *testing.T) {
	sut := newDecoder("retry: 100\nretry: a\n")

	_, err := sut.Decode()
	assert.ErrorIs(t, io.EOF, err)
	assert.Equal(t, 100*time.Millisecond, sut.Retry())
}

func BenchmarkDecodeEmptyEvent(b *testing.B) {
	runDecodingBenchmark(b, []byte("data: \n\n"))
}

func BenchmarkDecodeEmptyEventWithIgnoredLine(b *testing.B) {
	runDecodingBenchmark(b, []byte(":ignored line \n\ndata: \n\n"))
}

func BenchmarkDecodeShortEvent(b *testing.B) {
	runDecodingBenchmark(b, []byte("data: short event\n\n"))
}

func BenchmarkDecode1kEvent(b *testing.B) {
	runDecodingBenchmark(b, getBenchmarkPayload(1000))
}

func BenchmarkDecode4kEvent(b *testing.B) {
	runDecodingBenchmark(b, getBenchmarkPayload(4000))
}

func BenchmarkDecode8kEvent(b *testing.B) {
	runDecodingBenchmark(b, getBenchmarkPayload(8000))
}

func BenchmarkDecode16kEvent(b *testing.B) {
	runDecodingBenchmark(b, getBenchmarkPayload(16000))
}

func newDecoder(data string) *Decoder {
	reader := bytes.NewReader([]byte(data))
	return New(reader)
}

func runDecodingBenchmark(b *testing.B, data []byte) {
	reader := bytes.NewReader(data)
	decoder := New(reader)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoder.Decode()
		reader.Seek(0, 0)
	}
}

func getBenchmarkPayload(dataSize int) []byte {
	event := testutils.NewMessageEvent("event-id", "event-name", dataSize)
	out := new(bytes.Buffer)

	e := encoder.New(out)
	e.WriteComment("benchmark event")
	e.WriteRetry(2000)
	e.WriteEvent(event)
	return out.Bytes()
}
