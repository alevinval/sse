package base

var _ (MessageEventGetter) = (*MessageEvent)(nil)

// MessageEventGetter used by the decoder to be able to write any implementation
// of message event
type MessageEventGetter interface {
	GetName() string
	GetData() string
	GetLastEventID() string
}

// MessageEvent presents the payload being parsed from an EventSource.
type MessageEvent struct {
	Name        string
	Data        string
	LastEventID string
}

func (m *MessageEvent) GetName() string {
	return m.Name
}

func (m *MessageEvent) GetData() string {
	return m.Data
}

func (m *MessageEvent) GetLastEventID() string {
	return m.LastEventID
}
