package eventsource

// Status groups together a ready state and possible error associated
// with it. Useful to notify changes and why they happened.
type Status struct {
	ReadyState ReadyState
	Err        error
}

func (s *Status) Error() string {
	return s.Err.Error()
}
