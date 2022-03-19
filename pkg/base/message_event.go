package base

var _ (MessageEventGetter) = (*MessageEvent)(nil)

// MessageEventGetter used by the decoder to be able to write any implementation
// of message event
type MessageEventGetter interface {
	GetID() string
	GetName() string
	GetData() string
}

// MessageEvent presents the payload being parsed from an EventSource.
type MessageEvent struct {
	ID   string
	Name string
	Data string

	// HasID is used to signal that the ID has been reset.
	// This is necessary because we cannot differentiate empty string from
	// whether it was not sent, or it was sent with empty value.
	HasID bool
}

// GetID returns the ID of the event.
func (m *MessageEvent) GetID() string {
	return m.ID
}

// GetName returns the name of the event.
func (m *MessageEvent) GetName() string {
	return m.Name
}

// GetData returns the data of the event.
func (m *MessageEvent) GetData() string {
	return m.Data
}
