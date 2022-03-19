package testutils

import (
	"math/rand"

	"github.com/go-rfc/sse/pkg/base"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func NewMessageEvent(id, name string, dataSize int) *base.MessageEvent {
	return &base.MessageEvent{ID: id, Name: name, Data: randString(dataSize)}
}

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
