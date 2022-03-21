package base

import "github.com/go-rfc/sse/pkg/base/optional"

var _ (MessageEventGetter) = (*MessageEvent)(nil)

// MessageEventGetter used by the decoder to be able to write any implementation
// of message event
type MessageEventGetter interface {
	GetID() optional.Optional[string]
	GetName() (name string)
	GetData() (data string)
}

// MessageEvent presents the payload being parsed from an EventSource.
type MessageEvent struct {
	ID   optional.Optional[string]
	Name string
	Data string
}

// GetID returns the ID of the event.
func (m *MessageEvent) GetID() optional.Optional[string] {
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
