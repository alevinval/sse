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
	// DefaultDecoder is the decoder used by EventSource by default.
	DefaultDecoder = NewDecoder(defaultBufferSize)

	bytesLF    = []byte("\n")
	bytesCRLF  = []byte("\r\n")
	bytesSPACE = []byte(" ")
	bytesCOLON = []byte(":")
)

type (
	// Decoder interface decodes events from a reader input
	Decoder interface {
		Decode(in io.Reader) (out <-chan Event)
	}
	decoder struct {
		bufferSize int
	}
)

// NewDecoder builds an SSE decoder with the specified buffer size.
func NewDecoder(bufferSize int) Decoder {
	d := &decoder{}
	d.initialise(bufferSize)
	return d
}

func (d *decoder) initialise(bufferSize int) {
	d.bufferSize = bufferSize
}

// Returns a channel of SSE events from a reader input.
func (d *decoder) Decode(in io.Reader) <-chan Event {
	buffIn := bufio.NewReaderSize(in, d.bufferSize)
	out := make(chan Event)
	go process(buffIn, out)
	return out
}

// Processes a reader and sends the parsed SSE events
// to the output channel.
// This function is intended to run in a go-routine.
func process(in *bufio.Reader, out chan Event) {
	// Stores event data, which is filled after one or many lines from the reader
	var eventID, eventType, dataBuffer = new(bytes.Buffer), new(bytes.Buffer), new(bytes.Buffer)

	// Stores data about the current line being processed
	var field, value = new(bytes.Buffer), new(bytes.Buffer)

	for {
		line, err := in.ReadSlice('\n')
		if err != nil {
			close(out)
			return
		}

		// Dispatch event
		if bytes.Equal(line, bytesLF) || bytes.Equal(line, bytesCRLF) {
			// Skip event if Data buffer its empty
			if dataBuffer.Len() == 0 {
				dataBuffer.Reset()
				eventType.Reset()
				continue
			}

			data := dataBuffer.Bytes()

			// Trim last byte if line feed
			data = bytes.TrimSuffix(data, bytesLF)

			// Create event
			event := newEvent(eventID.String(), eventType.String(), data)

			// Clear event buffers
			eventType.Reset()
			dataBuffer.Reset()

			// Dispatch event
			out <- event
			continue
		}

		// Sanitise line feeds
		line = sanitiseLineFeed(line)

		// Extract field/value for current line
		field.Reset()
		value.Reset()

		colonIndex := bytes.Index(line, bytesCOLON)
		switch colonIndex {
		case 0:
			continue
		case -1:
			field.Write(line)
		default:
			field.Write(line[:colonIndex])
			line = line[colonIndex+1:]
			line = bytes.TrimPrefix(line, bytesSPACE)
			value.Write(line)
		}

		// Process field
		fieldName := field.String()
		switch fieldName {
		case "event":
			eventType.Write(value.Bytes())
		case "data":
			dataBuffer.Write(value.Bytes())
			dataBuffer.WriteByte('\n')
		case "id":
			eventID.Reset()
			eventID.Write(value.Bytes())
		case "retry":
			// TODO(alevinval): unused at the moment, will need refactor
			// or change on the internal API, as decoder has no knowledge on the underlying connection.
		default:
			// Ignore field
		}
	}
}

// Sanitises line feed ending.
func sanitiseLineFeed(line []byte) []byte {
	if bytes.HasSuffix(line, bytesCRLF) {
		return bytes.TrimSuffix(line, bytesCRLF)
	} else {
		return bytes.TrimSuffix(line, bytesLF)
	}
}
