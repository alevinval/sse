package testutils

import (
	"github.com/go-rfc/sse/pkg/base"
)

func NewMessageEvent(lastEventID, name string, dataSize int) *base.MessageEvent {
	data := make([]byte, dataSize)
	for i := range data {
		data[i] = 'e'
	}
	return &base.MessageEvent{LastEventID: lastEventID, Name: name, Data: string(data)}
}
