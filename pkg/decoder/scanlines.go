package decoder

// scanLinesCRLF is a variation of bufio.ScanLines that also recognizes
// just CR as EOL (as specified in the EventSource spec)
func scanLinesCR(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	for i, c := range data {
		switch c {
		case '\r':
			if i == len(data)-1 {
				if atEOF {
					return len(data), data[:i], nil
				}
				// We have to wait for the next byte to check if it is
				// a \n
				return 0, nil, nil
			}
			j := i
			i++
			if data[i] == '\n' {
				i++
			}
			return i, data[:j], nil
		case '\n':
			return i + 1, data[:i], nil
		}
	}

	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
