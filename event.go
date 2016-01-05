package sse

type (
	Event struct {
		id    string
		event string
		data  []byte
	}
)

func (me *Event) Id() string {
	return me.id
}

func (me *Event) Event() string {
	return me.event
}

func (me *Event) Data() []byte {
	return me.data[:]
}
