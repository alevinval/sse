package sse

import (
	"bufio"
	"bytes"
	"io"
)

// Returns a channel of SSE events from a reader input.
func Decode(reader io.Reader) chan *Event {
	output := make(chan *Event)
	go process(reader, output)
	return output
}

// Processes a reader and sends the parsed SSE events
// to the output channel.
// This function is intended to run in a go-routine
func process(reader io.Reader, out chan *Event) {
	in := bufio.NewReader(reader)

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
		if bytes.Equal(line, []byte("\r\n")) || bytes.Equal(line, []byte("\n")) || bytes.Equal(line, []byte("\r")) {
			if dataBuffer.Len() == 0 {
				dataBuffer.Reset()
				eventType.Reset()
				continue
			}

			// Trim last byte if its a line feed
			data := dataBuffer.Bytes()
			data = bytes.TrimSuffix(data, []byte("\n"))

			// Create event and reset buffers
			event := &Event{event: eventType.String(), data: make([]byte, len(data))}
			copy(event.data, data)

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

		// Extract field/value for current line
		field.Reset()
		value.Reset()
		colonIndex := bytes.Index(line, []byte(":"))
		if colonIndex != -1 {
			// Name
			field.Write(line[:colonIndex])

			// Value
			line = line[colonIndex+1:]
			line = bytes.TrimPrefix(line, []byte(" "))
			line = bytes.TrimRight(line, "\r\n")
			value.Write(line)
		} else {
			// Name
			line = bytes.TrimRight(line, " \r\n")
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
