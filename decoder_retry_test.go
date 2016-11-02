package sse

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeRetry(t *testing.T) {
	decoder := NewDecoder(bytes.NewReader([]byte("retry: 100\nretry: a\n")))
	ev, err := decoder.Decode()
	if assert.NoError(t, err) {
		assert.Equal(t, 100, ev.retry)
		_, err = decoder.Decode()
		assert.Equal(t, io.EOF, err)
	}
}
