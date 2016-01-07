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
		line = sanitiseLineEnding(line)

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

func sanitiseLineEnding(in []byte) []byte {
	var sanitisedLine []byte
	if bytes.HasSuffix(in, []byte("\r\n")) {
		sanitisedLine = bytes.TrimSuffix(in, []byte("\r\n"))
	} else {
		sanitisedLine = bytes.TrimSuffix(in, []byte("\n"))
	}
	return sanitisedLine
}
