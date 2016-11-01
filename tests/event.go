package tests

// NewEventWithPadding creates a raw slice of bytes with an event that does
// not exceed the specified size.
func NewEventWithPadding(size int) []byte {
	event := []byte("data: ")
	paddingByte := byte('e')
	for x := 0; x < size-8; x++ {
		event = append(event, paddingByte)
	}
	return append(event, []byte("\n\n")...)
}

// GetPaddedEventData returns the event data as it would be returned from
// .Data on the dispatched event.
func GetPaddedEventData(b []byte) string {
	return string(b[6 : len(b)-2])
}
