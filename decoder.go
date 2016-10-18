package sse

import (
	"bufio"
	"bytes"
	"io"
)

const (
	defaultBufferSize = 8096
)

const (
	byteLF    = '\n'
	byteCR    = '\r'
	byteSPACE = ' '
	byteCOLON = ':'
)

var (
	// DefaultDecoder is the decoder used by EventSource by default.
	DefaultDecoder = &decoder{defaultBufferSize}
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
		line, err := in.ReadSlice(byteLF)
		if err != nil {
			close(out)
			return
		}

		// Empty line? => Dispatch event
		if len(line) == 1 || (len(line) == 2 && line[0] == byteCR) {
			// Skip event if Data buffer is empty
			if dataBuffer.Len() == 0 {
				dataBuffer.Reset()
				eventType.Reset()
				continue
			}

			data := dataBuffer.Bytes()

			// Trim last byte if line feed
			if data[len(data)-1] == byteLF {
				data = data[:len(data)-1]
			}

			// Create event
			event := newEvent(eventID.String(), eventType.String(), data)

			// Clear event buffers
			eventType.Reset()
			dataBuffer.Reset()

			// Dispatch event
			out <- event
			continue
		}

		colonIndex := bytes.IndexByte(line, byteCOLON)

		// Sanitise line feeds
		line = sanitiseLineFeed(line)

		// Extract field/value for current line
		field.Reset()
		value.Reset()

		switch colonIndex {
		case 0:
			continue
		case -1:
			field.Write(line)
		default:
			field.Write(line[:colonIndex])
			line = line[colonIndex+1:]
			if len(line) > 0 && line[0] == byteSPACE {
				line = line[1:]
			}
			value.Write(line)
		}

		// Process field
		fieldName := field.String()
		switch fieldName {
		case "event":
			eventType.Write(value.Bytes())
		case "data":
			dataBuffer.Write(value.Bytes())
			dataBuffer.WriteByte(byteLF)
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
	l := len(line)
	// Trim LF
	if l > 0 && line[l-1] == byteLF {
		l--
		// Trim CR
		if l > 0 && line[l-1] == byteCR {
			l--
		}
		line = line[:l]
	}
	return line
}
