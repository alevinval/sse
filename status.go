package sse

// Status groups together a ready state and possible error associated
// with it. Useful to notify changes and why they happened.
type Status struct {
	ReadyState ReadyState
	Error      error
}
