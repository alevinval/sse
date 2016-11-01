package sse

type (
	Event struct {
		id    string
		name  string
		Retry int
		data  []byte
	}
)

func newEvent(id, name string, data []byte) *Event {
	e := &Event{}
	e.initialise(id, name, data)
	return e
}

func newRetryEvent(retry int) *Event {
	return &Event{Retry: retry}
}

// Initialises a new event struct.
// Performs a buffer allocation, and copies the data over.
func (e *Event) initialise(id, name string, data []byte) {
	e.id = id
	e.name = name
	e.data = make([]byte, len(data))
	e.Retry = -1
	copy(e.data, data)
}

func (e *Event) ID() string {
	return e.id
}

func (e *Event) Name() string {
	return e.name
}

func (e *Event) Data() []byte {
	return e.data
}
