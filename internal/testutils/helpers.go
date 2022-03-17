package testutils

import (
	"bytes"

	"github.com/go-rfc/sse/pkg/base"
	"github.com/go-rfc/sse/pkg/encoder"
)

func NewMessageEvent(lastEventID, name string, dataSize int) *base.MessageEvent {
	data := make([]byte, dataSize)
	for i := range data {
		data[i] = 'e'
	}
	return &base.MessageEvent{LastEventID: lastEventID, Name: name, Data: string(data)}
}

func NewMessageEventString(lastEventID, name string, dataSize int) string {
	ev := NewMessageEvent(lastEventID, name, dataSize)
	return MessageEventToString(ev)
}

func MessageEventToString(ev *base.MessageEvent) string {
	out := new(bytes.Buffer)
	e := encoder.New(out)
	e.Write(ev)
	return out.String()
}

func RetryEventToString(n int) string {
	out := new(bytes.Buffer)
	e := encoder.New(out)
	e.SetRetry(n)
	return out.String()
}
