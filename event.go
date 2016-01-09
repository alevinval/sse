package sse

type (
	Event struct {
		id    string
		event string
		data  []byte
	}
)

func NewEvent(id, event string, data []byte) Event {
	ev := Event{}
	ev.initialise(id, event, data)
	return ev
}

func (me *Event) initialise(id, event string, data []byte) {
	me.id = id
	me.event = event
	me.data = make([]byte, len(data))
	copy(me.data, data)
}

func (me *Event) Id() string {
	return me.id
}

func (me *Event) Event() string {
	return me.event
}

func (me *Event) Data() []byte {
	return me.data
}
