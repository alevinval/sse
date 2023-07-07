package testutils

import (
	"math/rand"
	"testing"
	"time"

	"github.com/alevinval/sse/pkg/base"
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

func ExpectCondition(t *testing.T, condition func() bool) {
	n, max, sleep := 0, 10, 25
	for n < max {
		n++
		if condition() {
			return
		}
		time.Sleep(time.Duration(sleep) * time.Millisecond)
	}

	t.Errorf("expected condition never happened after %d polls", n)
}
