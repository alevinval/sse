package sse

import (
	"fmt"
)

func newMessageEvent(lastEventID, name string, dataSize int) *MessageEvent {
	data := make([]byte, dataSize)
	for i := range data {
		data[i] = 'e'
	}
	return &MessageEvent{LastEventID: lastEventID, Name: name, Data: string(data)}
}

func newRetryEvent(ms int) string {
	return fmt.Sprintf("retry: %d\n", ms)
}

func messageEventToString(ev *MessageEvent) string {
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
