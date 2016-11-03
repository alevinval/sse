package sse

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"
)

// Default retry time in milliseconds.
// The spec recommends to use a value of a few seconds.
const defaultRetry = 2500

type (
	// Decoder interface decodes events from a reader input
	Decoder interface {
		Decode() (ev *MessageEvent, err error)
		Retry() (retry int)
	}
	decoder struct {
		lastEventID string
		retry       int
		scanner     *bufio.Scanner
		data        *bytes.Buffer
	}
)

func (d *decoder) Retry() int {
	return d.retry
}

// Decode reads the input stream and interprets the events in it. Any error while reading is  returned.
func (d *decoder) Decode() (*MessageEvent, error) {
	// Stores event data, which is filled after one or many lines from the reader
	var name string
	var eventSeen bool

	scanner := d.scanner
	data := d.data
	data.Reset()
	for scanner.Scan() {
		line := scanner.Text()
		// Empty line? => Dispatch event
		if len(line) == 0 {
			if eventSeen {
				// Trim the last LF
				if l := data.Len(); l > 0 {
					data.Truncate(l - 1)
				}
				// Note the event source spec as defined by w3.org requires
				// skips the event dispatching if the event name collides with
				// the name of any event as defined in the DOM Events spec.
				// Decoder does not perform this check, hence it could yield
				// events that would not be valid in a browser.
				return &MessageEvent{d.lastEventID, name, data.String()}, nil
			}
			continue
		}

		colonIndex := strings.IndexByte(line, ':')
		if colonIndex == 0 {
			// Skip comment
			continue
		}

		var fieldName, value string
		if colonIndex == -1 {
			fieldName = line
			value = ""
		} else {
			// Extract key/value for current line
			fieldName = line[:colonIndex]
			if colonIndex < len(line)-1 && line[colonIndex+1] == ' ' {
				// Trim prefix space
				value = line[colonIndex+2:]
			} else {
				value = line[colonIndex+1:]
			}
		}

		switch fieldName {
		case "event":
			name = value
			eventSeen = true
		case "data":
			data.WriteString(value)
			data.WriteByte('\n')
			eventSeen = true
		case "id":
			d.lastEventID = value
			eventSeen = true
		case "retry":
			retry, err := strconv.Atoi(value)
			if err == nil && retry >= 0 {
				d.retry = retry
			}

		default:
			// Ignore field
		}
	}

	// From the specification:
	// "Once the end of the file is reached, any pending data must be
	//  discarded. (If the file ends in the middle of an event, before the final
	//  empty line, the incomplete event is not dispatched.)"
	return nil, io.EOF
}
