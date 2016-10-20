package benchmarks_test

import (
	"bytes"
	"testing"

	"github.com/mubit/sse"
)

func runDecodingBenchmark(b *testing.B, data []byte) {
	reader := bytes.NewReader(data)
	decoder := sse.NewDecoder(reader)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoder.Decode()
		reader.Seek(0, 0)
	}
}

func BenchmarkDecodeEmptyEvent(b *testing.B) {
	event := []byte("data: \n\n")
	runDecodingBenchmark(b, event)
}

func BenchmarkDecodeEmptyEventWithIgnoredLine(b *testing.B) {
	event := []byte(":ignored line \n\ndata: \n\n")
	runDecodingBenchmark(b, event)
}

func BenchmarkDecodeShortEvent(b *testing.B) {
	event := []byte("data: short event\n\n")
	runDecodingBenchmark(b, event)
}

func createEventWithPadding(size int) []byte {
	event := []byte("data: ")
	paddingByte := byte('e')
	for x := 0; x < size-8; x++ {
		event = append(event, paddingByte)
	}
	return append(event, []byte("\n\n")...)
}

func BenchmarkDecode1kEvent(b *testing.B) {
	runDecodingBenchmark(b, createEventWithPadding(1000))
}

func BenchmarkDecode4kEvent(b *testing.B) {
	runDecodingBenchmark(b, createEventWithPadding(4000))
}

func BenchmarkDecode8kEvent(b *testing.B) {
	runDecodingBenchmark(b, createEventWithPadding(8000))
}

func BenchmarkDecode16kEvent(b *testing.B) {
	runDecodingBenchmark(b, createEventWithPadding(16000))
}
