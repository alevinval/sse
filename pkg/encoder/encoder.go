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

// WriteEvent encodes a full event.
//
// Note: this does not allow resetting the event source LastEventID, because
// empty IDs are simply ignored, and we cannot distinguish wether the
// intention of empty string was to reset the ID or to not send any.
// See WriteID.
func (e *Encoder) WriteEvent(event base.MessageEventGetter) (int, error) {
	e.buf.Reset()

	if event.GetID() != "" {
		e.buf.WriteString("id: ")
		e.buf.WriteString(event.GetID())
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

// WriteRetry encodes the retry field.
func (e *Encoder) WriteRetry(retryDelayInMillis int) {
	e.buf.Reset()
	e.buf.WriteString("retry: ")
	e.buf.WriteString(strconv.Itoa(retryDelayInMillis))
	e.buf.WriteByte('\n')
	e.out.Write(e.buf.Bytes())
}

// WriteID encodes an event id.
// This can be used to reset the LastEventID of a stream, since empty values
// on a MessageEvent are ignored, call WriteID to force sending an empty ID.
func (e *Encoder) WriteID(id string) {
	e.buf.Reset()
	e.buf.WriteString("id")
	if id != "" {
		e.buf.WriteString(": ")
		e.buf.WriteString(id)
	}
	e.buf.WriteByte('\n')
	e.out.Write(e.buf.Bytes())
}

// WriteComment encodes a comment. These are ignored by the decoder.
func (e *Encoder) WriteComment(comment string) {
	e.buf.Reset()
	e.buf.WriteByte(':')
	e.buf.WriteString(comment)
	e.buf.WriteByte('\n')
	e.out.Write(e.buf.Bytes())
}
