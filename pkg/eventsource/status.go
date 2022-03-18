package eventsource

// Status groups together a ready state and possible error associated
// with it. Useful to notify changes and why they happened.
type Status struct {
	Err        error
	ReadyState ReadyState
}

// Error returns the error that accompanies the ready-state change, if any
func (s *Status) Error() string {
	if s.Err != nil {
		return s.Err.Error()
	}
	return ""
}
