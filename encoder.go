package sse

import (
	"bytes"
	"fmt"
	"io"
)

type Encoder struct {
	buf *bytes.Buffer
	out io.Writer
}

func NewEncoder(out io.Writer) *Encoder {
	return &Encoder{
		buf: new(bytes.Buffer),
		out: out,
	}
}

func (e *Encoder) Write(event *MessageEvent) (int, error) {
	e.buf.Reset()

	if event.LastEventID != "" {
		e.buf.WriteString("id: " + event.LastEventID + "\n")
	}

	if event.Name != "" {
		e.buf.WriteString("name: " + event.Name + "\n")
	}

	if event.Data != "" {
		e.buf.WriteString("data: " + event.Data + "\n")
	}

	e.buf.WriteString("\n")

	return e.out.Write(e.buf.Bytes())
}

func (e *Encoder) SetRetry(retryDelayInMillis int) {
	e.buf.Reset()
	e.buf.WriteString(fmt.Sprintf("retry: %d\n", retryDelayInMillis))
	e.out.Write(e.buf.Bytes())
}
