package encoder

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"

	"github.com/alevinval/sse/internal"
	"github.com/alevinval/sse/pkg/base"
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
func (e *Encoder) WriteEvent(event base.MessageEventGetter) (int, error) {
	e.buf.Reset()

	if id, hasID := event.GetID(); id != "" || hasID {
		if id == "" {
			e.buf.WriteString("id\n")
		} else {
			e.buf.WriteString("id: ")
			e.buf.WriteString(id)
			e.buf.WriteByte('\n')
		}
	}

	if event.GetName() != "" {
		e.buf.WriteString("event: ")
		e.buf.WriteString(event.GetName())
		e.buf.WriteByte('\n')
	}

	if event.GetData() != "" {
		input := strings.NewReader(event.GetData())
		scanner := bufio.NewScanner(input)
		scanner.Split(internal.ScanLinesCR)
		for scanner.Scan() {
			line := scanner.Text()
			if len(line) == 0 {
				break
			}
			e.buf.WriteString("data: ")
			e.buf.WriteString(line)
			e.buf.WriteByte('\n')
		}
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

// WriteComment encodes a comment. These are ignored by the decoder.
func (e *Encoder) WriteComment(comment string) {
	e.buf.Reset()
	e.buf.WriteByte(':')
	e.buf.WriteString(comment)
	e.buf.WriteByte('\n')
	e.out.Write(e.buf.Bytes())
}
