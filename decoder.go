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
	DefaultDecoder = &decoder{defaultBufferSize}

	bytesLF    = []byte("\n")
	bytesCRLF  = []byte("\r\n")
	bytesSPACE = []byte(" ")
	bytesCOLON = []byte(":")
)

type (
	decoder struct {
		bufferSize int
	}
)

func NewDecoder(bufferSize int) *decoder {
	d := &decoder{}
	d.initialise(bufferSize)
	return d
}

func (me *decoder) initialise(bufferSize int) {
	me.bufferSize = bufferSize
}

// Returns a channel of SSE events from a reader input.
func (me *decoder) Decode(reader io.Reader) <-chan Event {
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
			event := newEvent("", eventType.String(), data)

			// Clear event buffers
			eventType.Reset()
			dataBuffer.Reset()

			// Dispatch event
			out <- event
			continue
		}

		colonIndex := bytes.Index(line, bytesCOLON)

		// Sanitise line feeds
		line = sanitise(line)

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
			line = bytes.TrimPrefix(line, bytesSPACE)
			value.Write(line)
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
	if bytes.HasSuffix(line, bytesCRLF) {
		line = bytes.TrimSuffix(line, bytesCRLF)
	} else {
		line = bytes.TrimSuffix(line, bytesLF)
	}
	return line
}
