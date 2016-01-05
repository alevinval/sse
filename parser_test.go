package sse

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOneLineDataParse(t *testing.T) {
	data := []byte("data: this is a test\r\n\r\n")
	reader := bytes.NewReader(data)
	events := parseStream(reader)
	ev := <-events
	assert.Equal(t, "this is a test", string(ev.Data))

	data = []byte("data: this is a test\r\n\n")
	reader = bytes.NewReader(data)
	events = parseStream(reader)
	ev = <-events
	assert.Equal(t, "this is a test", string(ev.Data))
}

func TestBasicParse(t *testing.T) {
	data := []byte("data: this is \r\ndata: a test\r\n\r\n")
	reader := bytes.NewReader(data)
	events := parseStream(reader)
	ev := <-events
	assert.Equal(t, "this is \na test", string(ev.Data))
}
