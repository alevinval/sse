package testutils

import "fmt"

// RetryEvent is used to represent a connection retry event
type RetryEvent struct {
	delayInMs int
}

// NewMessageEvent helper
func NewMessageEvent(lastEventID, name string, dataSize int) *MessageEvent {
	data := make([]byte, dataSize)
	for i := range data {
		data[i] = 'e'
	}
	return &MessageEvent{LastEventID: lastEventID, Name: name, Data: string(data)}
}

// NewRetryEvent helper
func NewRetryEvent(delayInMs int) *RetryEvent {
	return &RetryEvent{delayInMs}
}

// MessageEventToString serialiser
func MessageEventToString(ev *MessageEvent) string {
	msg := ""
	if ev.LastEventID != "" {
		msg = buildString("id: ", ev.LastEventID, "\n")
	}
	if ev.Name != "" {
		msg = buildString(msg, "event: ", ev.Name, "\n")
	}
	return buildString(msg, "data: ", ev.Data, "\n\n")
}

// RetryEventToString serialiser
func RetryEventToString(ev *RetryEvent) string {
	return buildString("retry: ", fmt.Sprintf("%d", ev.delayInMs), "\n")
}

func buildString(fields ...string) string {
	data := []byte{}
	for _, field := range fields {
		data = append(data, field...)
	}
	return string(data)
}
