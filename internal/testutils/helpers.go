package testutils

import (
	"math/rand"

	"github.com/go-rfc/sse/pkg/base"
	"github.com/go-rfc/sse/pkg/base/optional"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func NewMessageEvent(id, name string, dataSize int) *base.MessageEvent {
	return &base.MessageEvent{ID: optional.Of(id), Name: name, Data: randString(dataSize)}
}

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
