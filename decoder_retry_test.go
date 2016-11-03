package sse

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeRetry(t *testing.T) {
	decoder := NewDecoder(bytes.NewReader([]byte("retry: 100\nretry: a\n")))
	_, err := decoder.Decode()
	assert.Equal(t, 100, decoder.Retry())
	assert.Equal(t, io.EOF, err)
}
