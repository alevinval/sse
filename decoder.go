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

type (
	// Decoder interface decodes events from a reader input
	Decoder interface {
		Decode() (Event, error)
	}
	decoder struct {
		r *bufio.Reader

		// Buffers
		eventID    *bytes.Buffer
		eventType  *bytes.Buffer
		dataBuffer *bytes.Buffer
		field      *bytes.Buffer
		value      *bytes.Buffer
	}
)

// NewDecoder creates a SSE decoder with the default buffer size.
func NewDecoder(in io.Reader) Decoder {
	return NewDecoderSize(in, defaultBufferSize)
}

// NewDecoderSize creates a SSE decoder with the specified buffer size.
func NewDecoderSize(r io.Reader, bufferSize int) Decoder {
	d := new(decoder)
	d.initialise(r, bufferSize)
	return d
}

func (d *decoder) initialise(r io.Reader, bufferSize int) {
	normalizer := lineFeedNormalizer{r: r}
	d.r = bufio.NewReaderSize(&normalizer, bufferSize)

	// Event buffers
	d.eventID = new(bytes.Buffer)
	d.eventType = new(bytes.Buffer)
	d.dataBuffer = new(bytes.Buffer)

	// Parsing buffers
	d.field = new(bytes.Buffer)
	d.value = new(bytes.Buffer)
}

// Decode reads the input stream and interprets the events in it. Any error while reading is  returned.
func (d *decoder) Decode() (Event, error) {
	for {
		line, err := d.r.ReadBytes(byteLF)
		if err != nil {
			return nil, err
		}

		// Empty line? => Dispatch event
		// Note the event source spec as defined by w3.org requires skips the event dispatching if
		// the event name collides with the name of any event as defined in the DOM Events spec.
		// Decoder does not perform this check, hence it could yield events that would not be valid
		// in a browser.
		if len(line) == 1 {
			// Skip event if Data buffer is empty
			if d.dataBuffer.Len() == 0 {
				d.dataBuffer.Reset()
				d.eventType.Reset()
				continue
			}

			data := d.dataBuffer.Bytes()

			// Remove line feed, bounds already checked.
			data = unsafeTrimSuffixByte(data, byteLF)

			// Create event
			event := newEvent(d.eventID.String(), d.eventType.String(), data)

			// Clear event buffers
			d.eventType.Reset()
			d.dataBuffer.Reset()

			// Dispatch event
			return event, nil
		}

		// Remove line feed, bounds already checked.
		line = unsafeTrimSuffixByte(line, byteLF)

		// Extract field/value for current line
		d.field.Reset()
		d.value.Reset()

		colonIndex := bytes.IndexByte(line, byteCOLON)
		switch colonIndex {
		case 0:
			continue
		case -1:
			d.field.Write(line)
		default:
			d.field.Write(line[:colonIndex])
			line = line[colonIndex+1:]
			line = trimPrefixByte(line, byteSPACE)
			d.value.Write(line)
		}

		// Process field
		fieldName := d.field.String()
		switch fieldName {
		case "event":
			d.eventType.Write(d.value.Bytes())
		case "data":
			d.dataBuffer.Write(d.value.Bytes())
			d.dataBuffer.WriteByte(byteLF)
		case "id":
			d.eventID.Reset()
			d.eventID.Write(d.value.Bytes())
		case "retry":
			// TODO(alevinval): unused at the moment, will need refactor
			// or change on the internal API, as decoder has no knowledge on the underlying connection.
		default:
			// Ignore field
		}
	}
}

type lineFeedNormalizer struct {
	r    io.Reader
	last byte
}

func (lnf *lineFeedNormalizer) Read(p []byte) (int, error) {
	n, err := lnf.r.Read(p)
	for i := 0; i < n; i++ {
		switch p[i] {
		case byteLF:
			if lnf.last == byteCR {
				lnf.last = byteLF
				copy(p[i:], p[i+1:])
				n--
				i--
			}
		case byteCR:
			lnf.last = byteCR
			p[i] = byteLF
		default:
			lnf.last = p[i]
		}
	}
	return n, err
}

func trimPrefixByte(b []byte, prefix byte) []byte {
	if len(b) > 0 && b[0] == prefix {
		return b[1:]
	}
	return b
}

// Trims a suffix without doing a bounds check.
func unsafeTrimSuffixByte(b []byte, suffix byte) []byte {
	l := len(b) - 1
	if b[l] == suffix {
		return b[:l]
	}
	return b
}
