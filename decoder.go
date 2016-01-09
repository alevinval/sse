package sse

import (
	"bufio"
	"bytes"
	"io"
)

const (
	defaultBufferSize = 8096
)

var (
	DefaultDecoder = &Decoder{defaultBufferSize}
)

type (
	Decoder struct {
		bufferSize int
	}
)

func NewDecoder(bufferSize int) *Decoder {
	d := &Decoder{}
	d.initialise(bufferSize)
	return d
}

// Default decode function, with default buffer size
func Decode(reader io.Reader) <-chan Event {
	return DefaultDecoder.Decode(reader)
}

func (me *Decoder) initialise(bufferSize int) {
	me.bufferSize = bufferSize
}

// Returns a channel of SSE events from a reader input.
func (me *Decoder) Decode(reader io.Reader) <-chan Event {
	in := bufio.NewReaderSize(reader, me.bufferSize)
	out := make(chan Event)
	go process(in, out)
	return out
}

// Processes a reader and sends the parsed SSE events
// to the output channel.
// This function is intended to run in a go-routine.
func process(in *bufio.Reader, out chan Event) {
	// Stores event data, which is filled after one or many lines from the reader
	var eventType, dataBuffer = new(bytes.Buffer), new(bytes.Buffer)

	// Stores data about the current line being processed
	var field, value = new(bytes.Buffer), new(bytes.Buffer)

	for {
		line, err := in.ReadSlice('\n')
		if err != nil {
			close(out)
			return
		}

		// Dispatch event
		if bytes.Equal(line, []byte("\n")) || bytes.Equal(line, []byte("\r\n")) {
			// Skip event if Data buffer its empty
			if dataBuffer.Len() == 0 {
				dataBuffer.Reset()
				eventType.Reset()
				continue
			}

			data := dataBuffer.Bytes()

			// Trim last byte if line feed
			data = bytes.TrimSuffix(data, []byte("\n"))

			// Create event
			event := NewEvent("", eventType.String(), data)

			// Clear event buffers
			eventType.Reset()
			dataBuffer.Reset()

			// Dispatch event
			out <- event
			continue
		}

		// Ignore line
		if line[0] == ':' {
			continue
		}

		// Sanitise line feeds
		line = sanitise(line)

		// Extract field/value for current line
		field.Reset()
		value.Reset()
		colonIndex := bytes.Index(line, []byte(":"))
		if colonIndex != -1 {
			field.Write(line[:colonIndex])
			line = line[colonIndex+1:]
			line = bytes.TrimPrefix(line, []byte(" "))
			value.Write(line)
		} else {
			field.Write(line)
		}

		// Process field
		fieldName := field.String()
		switch fieldName {
		case "event":
			eventType.WriteString(fieldName)
		case "data":
			dataBuffer.Write(value.Bytes())
			dataBuffer.WriteByte('\n')
		case "id", "retry":
			// TODO(alevinval): unused at the moment, together with reconnection time
		default:
			// Ignore field
		}
	}
}

// Sanitises line feed ending.
func sanitise(line []byte) []byte {
	if bytes.HasSuffix(line, []byte("\r\n")) {
		line = bytes.TrimSuffix(line, []byte("\r\n"))
	} else {
		line = bytes.TrimSuffix(line, []byte("\n"))
	}
	return line
}
