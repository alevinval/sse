package sse

type (
	Event interface {
		Id() string
		Name() string
		Data() []byte
	}
	event struct {
		id   string
		name string
		data []byte
	}
)

func NewEvent(id, name string, data []byte) Event {
	e := &event{}
	e.initialise(id, name, data)
	return e
}

// Initialises a new event struct.
// Performs a buffer allocation, and copies the data over.
func (me *event) initialise(id, name string, data []byte) {
	me.id = id
	me.name = name
	me.data = make([]byte, len(data))
	copy(me.data, data)
}

func (me *event) Id() string {
	return me.id
}

func (me *event) Name() string {
	return me.name
}

func (me *event) Data() []byte {
	return me.data
}
