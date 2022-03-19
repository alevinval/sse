package testutils

import (
	"github.com/go-rfc/sse/pkg/base"
)

func NewMessageEvent(id, name string, dataSize int) *base.MessageEvent {
	data := make([]byte, dataSize)
	for i := range data {
		data[i] = 'e'
	}
	return &base.MessageEvent{ID: id, Name: name, Data: string(data)}
}
