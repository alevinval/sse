package sse

type (
	// Event is the interface that all events must satisfy
	Event interface {
		ID() (id string)
		Name() (name string)
		Data() (data []byte)
	}
	event struct {
		id   string
		name string
		data []byte
	}
)

func newEvent(id, name string, data []byte) *event {
	e := &event{}
	e.initialise(id, name, data)
	return e
}

// Initialises a new event struct.
// Performs a buffer allocation, and copies the data over.
func (e *event) initialise(id, name string, data []byte) {
	e.id = id
	e.name = name
	e.data = make([]byte, len(data))
	copy(e.data, data)
}

func (e *event) ID() string {
	return e.id
}

func (e *event) Name() string {
	return e.name
}

func (e *event) Data() []byte {
	return e.data
}
