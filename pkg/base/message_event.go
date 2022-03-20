package base

var _ (MessageEventGetter) = (*MessageEvent)(nil)

// MessageEventGetter used by the decoder to be able to write any implementation
// of message event
type MessageEventGetter interface {
	GetID() (id string, hasID bool)
	GetName() (name string)
	GetData() (data string)
}

// MessageEvent presents the payload being parsed from an EventSource.
type MessageEvent struct {
	ID   string
	Name string
	Data string

	// HasID is used to signal the ID has been explicitly set.
	// It is not necessary to enable this flag if ID != ""
	// It must be enabled when ID == ""
	HasID bool
}

// GetID returns the ID of the event.
func (m *MessageEvent) GetID() (id string, hasID bool) {
	return m.ID, m.HasID
}

// GetName returns the name of the event.
func (m *MessageEvent) GetName() string {
	return m.Name
}

// GetData returns the data of the event.
func (m *MessageEvent) GetData() string {
	return m.Data
}
