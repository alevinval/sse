package sse

type MessageEvent struct {
	LastEventID string
	Name        string
	Data        string
}
