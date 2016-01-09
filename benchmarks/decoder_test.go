package benchmarks_test

import (
	"bytes"
	"github.com/mubit/sse"
	"testing"
	"time"
)

func doBenchmark(b *testing.B, data []byte) {
	data = append(data, data...)
	reader := bytes.NewReader(data)
	decoder := sse.NewDecoder(65536)
	events := decoder.Decode(reader)
	time.Sleep(100 * time.Millisecond)

	for i := 0; i < b.N; i++ {
		reader.Seek(0, 0)
		<-events
	}
}

func BenchmarkDecodeEmptyEvent(b *testing.B) {
	b.ReportAllocs()
	event := []byte("data: \n\n")
	doBenchmark(b, event)
}

func BenchmarkDecodeEmptyEventWithIgnoredLine(b *testing.B) {
	b.ReportAllocs()
	event := []byte(":ignored line \n\ndata: \n\n")
	doBenchmark(b, event)
}

func BenchmarkDecodeShortEvent(b *testing.B) {
	b.ReportAllocs()
	event := []byte("data: short event\n\n")
	doBenchmark(b, event)
}

func BenchmarkDecode8kEvent(b *testing.B) {
	b.ReportAllocs()
	event := []byte("data: ")
	for x := 0; x < 8192; x++ {
		event = append(event, []byte("e")...)
	}
	event = append(event, []byte("\n\n")...)
	doBenchmark(b, event)
}

func BenchmarkDecode16kEvent(b *testing.B) {
	b.ReportAllocs()
	event := []byte("data: ")
	for x := 0; x < 16384; x++ {
		event = append(event, []byte("e")...)
	}
	event = append(event, []byte("\n\n")...)
	doBenchmark(b, event)
}
