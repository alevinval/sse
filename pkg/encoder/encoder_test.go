package encoder

import (
	"bytes"
	"testing"

	"github.com/go-rfc/sse/internal/testutils"
	"github.com/go-rfc/sse/pkg/base"
	"github.com/stretchr/testify/assert"
)

var (
	eventName      = &base.MessageEvent{Name: "first"}
	eventNameAndID = &base.MessageEvent{Name: "first", LastEventID: "1"}
	eventFull      = &base.MessageEvent{Name: "first", LastEventID: "1", Data: "some event data"}
)

func TestEncoderName(t *testing.T) {
	e, out := getEncoderAndOut()
	e.Write(eventName)
	assert.Equal(t, "name: first\n\n", out.String())
}

func TestEncoderNameAndID(t *testing.T) {
	e, out := getEncoderAndOut()
	e.Write(eventNameAndID)
	assert.Equal(t, "id: 1\nname: first\n\n", out.String())
}

func TestEncoderFullEvent(t *testing.T) {
	e, out := getEncoderAndOut()
	e.Write(eventFull)
	assert.Equal(t, "id: 1\nname: first\ndata: some event data\n\n", out.String())
}

func TestEncoderSetRetry(t *testing.T) {
	e, out := getEncoderAndOut()
	e.SetRetry(123)
	assert.Equal(t, "retry: 123\n", out.String())
}

func BenchmarkEncodeEmptyEvent(b *testing.B) {
	ev := &base.MessageEvent{}
	runEncodingBenchmark(b, ev)
}

func BenchmarkEncodeShortEvent(b *testing.B) {
	ev := &base.MessageEvent{Data: "some relatively short event"}
	runEncodingBenchmark(b, ev)
}

func BenchmarkEncode1kEvent(b *testing.B) {
	ev := testutils.NewMessageEvent("", "", 1000)
	runEncodingBenchmark(b, ev)
}

func BenchmarkEncode4kEvent(b *testing.B) {
	ev := testutils.NewMessageEvent("", "", 4000)
	runEncodingBenchmark(b, ev)
}

func BenchmarkEncode8kEvent(b *testing.B) {
	ev := testutils.NewMessageEvent("", "", 8000)
	runEncodingBenchmark(b, ev)
}

func BenchmarkEncode16kEvent(b *testing.B) {
	ev := testutils.NewMessageEvent("", "", 16000)
	runEncodingBenchmark(b, ev)
}

func getEncoderAndOut() (*Encoder, *bytes.Buffer) {
	out := new(bytes.Buffer)
	e := New(out)
	return e, out
}

func runEncodingBenchmark(b *testing.B, event *base.MessageEvent) {
	out := new(bytes.Buffer)
	encoder := New(out)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoder.Write(event)
		out.Reset()
	}
}
