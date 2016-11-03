package sse

type Event struct {
	LastEventID string
	Name        string
	Data        string
}
