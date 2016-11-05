package tests

import (
	"fmt"

	"github.com/mubit/sse"
)

// NewEventWithPadding creates a raw slice of bytes with an event that does
// not exceed the specified size.
func NewMessageEvent(lastEventID, name string, dataSize int) *sse.MessageEvent {
	data := make([]byte, dataSize)
	for i := range data {
		data[i] = 'e'
	}
	return &sse.MessageEvent{LastEventID: lastEventID, Name: name, Data: string(data)}
}

func NewRetryEvent(ms int) string {
	return fmt.Sprintf("retry: %d\n", ms)
}

// MessageEventToString encodes sse.MessageEvent into a string.
func MessageEventToString(ev *sse.MessageEvent) string {
	data := []byte{}
	if ev.LastEventID != "" {
		data = append(data, "id: "...)
		data = append(data, ev.LastEventID...)
		data = append(data, "\n"...)
	}
	if ev.Name != "" {
		data = append(data, "event: "...)
		data = append(data, ev.Name...)
		data = append(data, "\n"...)
	}
	data = append(data, "data: "...)
	data = append(data, ev.Data...)
	data = append(data, "\n\n"...)
	return string(data)
}
