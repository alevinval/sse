package sse

// MessageEvent presents the payload being parsed from an EventSource.
type MessageEvent struct {
	LastEventID string
	Name        string
	Data        string
}
