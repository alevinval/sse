package tests

func NewEventWithPadding(size int) []byte {
	event := []byte("data: ")
	paddingByte := byte('e')
	for x := 0; x < size-8; x++ {
		event = append(event, paddingByte)
	}
	return append(event, []byte("\n\n")...)
}
