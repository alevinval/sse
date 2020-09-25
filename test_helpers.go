package sse

import "bytes"

func newMessageEvent(lastEventID, name string, dataSize int) *MessageEvent {
	data := make([]byte, dataSize)
	for i := range data {
		data[i] = 'e'
	}
	return &MessageEvent{LastEventID: lastEventID, Name: name, Data: string(data)}
}

func newMessageEventString(lastEventID, name string, dataSize int) string {
	ev := newMessageEvent(lastEventID, name, dataSize)
	return messageEventToString(ev)
}

func messageEventToString(ev *MessageEvent) string {
	out := new(bytes.Buffer)
	e := NewEncoder(out)
	e.Write(ev)
	return out.String()
}

func retryEventToString(n int) string {
	out := new(bytes.Buffer)
	e := NewEncoder(out)
	e.SetRetry(n)
	return out.String()
}
