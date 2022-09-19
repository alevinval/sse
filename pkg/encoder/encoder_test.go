package encoder

import (
	"bytes"
	"testing"

	"github.com/alevinval/sse/internal/testutils"
	"github.com/alevinval/sse/pkg/base"
	"github.com/stretchr/testify/assert"
)

func TestEncoder_WriteEvent_EncodesName(t *testing.T) {
	event := &base.MessageEvent{Name: "event-name"}
	sut, out := getEncoder()

	sut.WriteEvent(event)

	assert.Equal(t, "event: event-name\n\n", out.String())
}

func TestEncoder_WriteEvent_EncodesIDWhenHasIDIsFalse(t *testing.T) {
	event := &base.MessageEvent{ID: "event-id"}
	sut, out := getEncoder()

	sut.WriteEvent(event)

	assert.Equal(t, "id: event-id\n\n", out.String())
}

func TestEncoder_WriteEvent_EncodesIDWhenHasIDIsTrue(t *testing.T) {
	event := &base.MessageEvent{ID: "event-id", HasID: true}
	sut, out := getEncoder()

	sut.WriteEvent(event)

	assert.Equal(t, "id: event-id\n\n", out.String())
}

func TestEncoder_WriteEvent_EncodesEmptyID(t *testing.T) {
	event := &base.MessageEvent{HasID: true}
	sut, out := getEncoder()

	sut.WriteEvent(event)

	assert.Equal(t, "id\n\n", out.String())
}

func TestEncoder_WriteEvent_EncodesData(t *testing.T) {
	event := &base.MessageEvent{Data: "event-data"}
	sut, out := getEncoder()

	sut.WriteEvent(event)

	assert.Equal(t, "data: event-data\n\n", out.String())
}

func TestEncoder_WriteEvent_EncodesFullEvent(t *testing.T) {
	event := &base.MessageEvent{ID: "event-id", Name: "event-name", Data: "event-data"}
	sut, out := getEncoder()

	sut.WriteEvent(event)

	assert.Equal(t, "id: event-id\nevent: event-name\ndata: event-data\n\n", out.String())
}

func TestEncoder_WriteRetry_EncodesRetry(t *testing.T) {
	sut, out := getEncoder()

	sut.WriteRetry(123)

	assert.Equal(t, "retry: 123\n", out.String())
}

func TestEncoder_WriteComment_EncodesCommentary(t *testing.T) {
	sut, out := getEncoder()

	sut.WriteComment("this is a commentary")

	assert.Equal(t, ":this is a commentary\n", out.String())
}

func getEncoder() (*Encoder, *bytes.Buffer) {
	out := new(bytes.Buffer)
	sut := New(out)
	return sut, out
}

func BenchmarkEncodeEmptyEvent(b *testing.B) {
	runEncodingBenchmark(b, 0)
}

func BenchmarkEncode128Event(b *testing.B) {
	runEncodingBenchmark(b, 128)
}

func BenchmarkEncode256Event(b *testing.B) {
	runEncodingBenchmark(b, 256)
}

func BenchmarkEncode512Event(b *testing.B) {
	runEncodingBenchmark(b, 512)
}

func BenchmarkEncode1kEvent(b *testing.B) {
	runEncodingBenchmark(b, 1024)
}

func BenchmarkEncode2kEvent(b *testing.B) {
	runEncodingBenchmark(b, 2048)
}

func runEncodingBenchmark(b *testing.B, dataSize int) {
	event := testutils.NewMessageEvent("event-id", "event-name", dataSize)
	out := new(bytes.Buffer)
	encoder := New(out)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoder.WriteComment("benchmark event")
		encoder.WriteRetry(2000)
		encoder.WriteEvent(event)
		out.Reset()
	}
}
