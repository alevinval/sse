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
		e.buf.WriteString("id: ")
		e.buf.WriteString(event.GetLastEventID())
		e.buf.WriteByte('\n')
	}

	if event.GetName() != "" {
		e.buf.WriteString("name: ")
		e.buf.WriteString(event.GetName())
		e.buf.WriteByte('\n')
	}

	if event.GetData() != "" {
		e.buf.WriteString("data: ")
		e.buf.WriteString(event.GetData())
		e.buf.WriteByte('\n')
	}

	e.buf.WriteByte('\n')

	return e.out.Write(e.buf.Bytes())
}

func (e *Encoder) SetRetry(retryDelayInMillis int) {
	e.buf.Reset()
	e.buf.WriteString("retry: ")
	e.buf.WriteString(strconv.Itoa(retryDelayInMillis))
	e.buf.WriteByte('\n')
	e.out.Write(e.buf.Bytes())
}
