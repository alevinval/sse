package eventsource

//go:generate stringer -type=ReadyState

// ReadyState indicates the state of the EventSource.
type ReadyState uint16

const (
	// Connecting while trying to establish connection with the stream.
	Connecting ReadyState = iota
	// Open after connection is established with the server.
	Open
	// Closing after Close is invoked.
	Closing
	// Closed after the connection is closed.
	Closed
)
