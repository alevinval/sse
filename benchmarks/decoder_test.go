package benchmarks_test

import (
	"bytes"
	"testing"

	"github.com/mubit/sse"
	"github.com/mubit/sse/tests"
)

func runDecodingBenchmark(b *testing.B, data string) {
	reader := bytes.NewReader([]byte(data))
	b.ResetTimer()
	decoder := sse.NewDecoder(reader)
	for i := 0; i < b.N; i++ {
		decoder.Decode()
		reader.Seek(0, 0)
	}
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
	ev := tests.NewMessageEvent("", "", 1000)
	runDecodingBenchmark(b, tests.MessageEventToString(ev))
}

func BenchmarkDecode4kEvent(b *testing.B) {
	ev := tests.NewMessageEvent("", "", 4000)
	runDecodingBenchmark(b, tests.MessageEventToString(ev))
}

func BenchmarkDecode8kEvent(b *testing.B) {
	ev := tests.NewMessageEvent("", "", 8000)
	runDecodingBenchmark(b, tests.MessageEventToString(ev))
}

func BenchmarkDecode16kEvent(b *testing.B) {
	ev := tests.NewMessageEvent("", "", 16000)
	runDecodingBenchmark(b, tests.MessageEventToString(ev))
}
