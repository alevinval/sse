package encoder

import (
	"bytes"
	"io"
	"strconv"

	"github.com/go-rfc/sse/pkg/base"
)

type Encoder struct {
	buf *bytes.Buffer
	out io.Writer
}

func New(out io.Writer) *Encoder {
	return &Encoder{
		buf: new(bytes.Buffer),
		out: out,
	}
}

func (e *Encoder) Write(event base.MessageEventGetter) (int, error) {
	e.buf.Reset()

	if event.GetLastEventID() != "" {
		e.buf.WriteString("id: " + event.GetLastEventID() + "\n")
	}

	if event.GetName() != "" {
		e.buf.WriteString("name: " + event.GetName() + "\n")
	}

	if event.GetData() != "" {
		e.buf.WriteString("data: " + event.GetData() + "\n")
	}

	e.buf.WriteString("\n")

	return e.out.Write(e.buf.Bytes())
}

func (e *Encoder) SetRetry(retryDelayInMillis int) {
	e.buf.Reset()
	e.buf.WriteString("retry: " + strconv.Itoa(retryDelayInMillis) + "\n")
	e.out.Write(e.buf.Bytes())
}
