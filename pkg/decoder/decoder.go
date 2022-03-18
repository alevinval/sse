package decoder

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/go-rfc/sse/pkg/base"
)

// Default retry time in milliseconds.
// The spec recommends to use a value of a few seconds.
const defaultRetry = time.Duration(2500) * time.Millisecond

type (
	// Decoder accepts an io.Reader input and decodes message events from it.
	Decoder struct {
		lastEventID string
		retry       time.Duration
		scanner     *bufio.Scanner
		data        *bytes.Buffer
	}
)

// New returns a Decoder with a growing buffer.
// Lines are limited to bufio.MaxScanTokenSize - 1.
func New(in io.Reader) *Decoder {
	return NewSize(in, 0)
}

// NewSize returns a Decoder with a fixed buffer size.
func NewSize(in io.Reader, bufferSize int) *Decoder {
	d := &Decoder{scanner: bufio.NewScanner(in), data: new(bytes.Buffer), retry: defaultRetry}
	if bufferSize > 0 {
		d.scanner.Buffer(make([]byte, bufferSize), bufferSize)
	}
	d.scanner.Split(scanLinesCR) // See scanlines.go
	return d
}

// Retry returns the to wait before attempting to reconnect to the event source.
func (d *Decoder) Retry() time.Duration {
	return d.retry
}

// Decode reads the input stream and parses events from it. Any error while reading is  returned.
func (d *Decoder) Decode() (*base.MessageEvent, error) {
	// Stores event data, which is filled after one or many lines from the reader
	var name string
	var eventSeen bool

	d.data.Reset()
	for d.scanner.Scan() {
		line := d.scanner.Text()
		// Empty line? => Dispatch event
		if len(line) == 0 {
			if eventSeen {
				// Trim the last LF
				if l := d.data.Len(); l > 0 {
					d.data.Truncate(l - 1)
				}

				// Note the event source spec as defined by w3.org requires
				// skips the event dispatching if the event name collides with
				// the name of any event as defined in the DOM Events spec.
				// Decoder does not perform this check, hence it could yield
				// events that would not be valid in a browser.
				return &base.MessageEvent{LastEventID: d.lastEventID, Name: name, Data: d.data.String()}, nil
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
			d.data.WriteString(value)
			d.data.WriteByte('\n')
			eventSeen = true
		case "id":
			d.lastEventID = value
			eventSeen = true
		case "retry":
			retry, err := strconv.Atoi(value)
			if err == nil && retry >= 0 {
				d.retry = time.Duration(retry) * time.Millisecond
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
