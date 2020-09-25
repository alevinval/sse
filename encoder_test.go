package sse

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	eventName      = &MessageEvent{Name: "first"}
	eventNameAndID = &MessageEvent{Name: "first", LastEventID: "1"}
	eventFull      = &MessageEvent{Name: "first", LastEventID: "1", Data: "some event data"}
)

func TestEncoderName(t *testing.T) {
	e, out := getEncoderAndOut()
	e.Write(eventName)
	assert.Equal(t, "name: first\n\n", out.String())
}

func TestEncoderNameAndID(t *testing.T) {
	e, out := getEncoderAndOut()
	e.Write(eventNameAndID)
	assert.Equal(t, "id: 1\nname: first\n\n", out.String())
}

func TestEncoderFullEvent(t *testing.T) {
	e, out := getEncoderAndOut()
	e.Write(eventFull)
	assert.Equal(t, "id: 1\nname: first\ndata: some event data\n\n", out.String())
}

func TestEncoderSetRetry(t *testing.T) {
	e, out := getEncoderAndOut()
	e.SetRetry(123)
	assert.Equal(t, "retry: 123\n", out.String())
}

func getEncoderAndOut() (*Encoder, *bytes.Buffer) {
	out := new(bytes.Buffer)
	e := NewEncoder(out)
	return e, out
}
