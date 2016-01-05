package sse

import (
	"bufio"
	"bytes"
	"io"
)

// Returns a channel of SSE events from a reader input.
func parseStream(reader io.Reader) <-chan Event {
	output := make(chan Event)
	go processReader(reader, output)
	return output
}

// Processes a reader and sends the parsed SSE events
// to the output channel.
// This function is intended to run in a go-routine
func processReader(reader io.Reader, out chan Event) {
	in := bufio.NewReader(reader)

	// Stores event data, which is filled after one or many lines from the reader
	var eventName, dataBuffer = "", new(bytes.Buffer)

	// Stores data about the current line being processed
	var fieldName, fieldValue = "", new(bytes.Buffer)

	for {
		line, err := in.ReadBytes('\n')
		if err != nil {
			continue
		}

		// Dispatch event
		if bytes.Equal(line, []byte("\r\n")) || bytes.Equal(line, []byte("\n")) || bytes.Equal(line, []byte("\r")) {
			out <- Event{Type: eventName, Data: []byte(dataBuffer.String())}
			fieldName, eventName = "", ""
			dataBuffer.Reset()
			fieldValue.Reset()
			continue
		}

		// Ignore line
		if line[0] == ':' {
			continue
		}

		// Extract field/value
		colonIndex := bytes.Index(line, []byte(":"))
		if colonIndex != -1 {
			fieldName = string(line[:colonIndex])
			value := line[colonIndex+1:]
			value = bytes.TrimLeft(value, " ")
			value = bytes.TrimRight(value, "\r\n")
			fieldValue.Reset()
			fieldValue.Write(value)
		} else {
			fieldName = string(line)
			fieldValue.Reset()
		}

		// Fill data buffer
		switch fieldName {
		case "event":
			eventName = fieldValue.String()
			break
		case "data":
			if dataBuffer.Len() > 0 {
				dataBuffer.WriteByte('\n')
			}
			dataBuffer.Write(fieldValue.Bytes())
			break
		case "id":
			// TODO(alevinval): unused at the moment, together with reconnection time
			break
		case "retry":
			// TODO(alevinval): don't ignore reconnection time?
			break
		default:
			// Ignore field
			break
		}
	}
}
