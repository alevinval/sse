package benchmarks_test

import (
	"bytes"
	"testing"

	"github.com/mubit/sse"
	"github.com/mubit/sse/tests"
)

func runDecodingBenchmark(b *testing.B, data []byte) {
	reader := bytes.NewReader(data)
	b.ResetTimer()
	decoder := sse.NewDecoder(reader)
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

func BenchmarkDecode1kEvent(b *testing.B) {
	runDecodingBenchmark(b, tests.NewEventWithPadding(1000))
}

func BenchmarkDecode4kEvent(b *testing.B) {
	runDecodingBenchmark(b, tests.NewEventWithPadding(4000))
}

func BenchmarkDecode8kEvent(b *testing.B) {
	runDecodingBenchmark(b, tests.NewEventWithPadding(8000))
}

func BenchmarkDecode16kEvent(b *testing.B) {
	runDecodingBenchmark(b, tests.NewEventWithPadding(16000))
}
