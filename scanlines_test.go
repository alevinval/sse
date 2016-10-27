package sse

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanLines(t *testing.T) {
	assert := assert.New(t)
	for _, test := range []struct {
		in      string
		atEOF   bool
		advance int
		line    string
		err     error
	}{
		{"", false, 0, "", nil},
		{"", true, 0, "", nil},
		{"\n", false, 1, "", nil},
		{"\n", true, 1, "", nil},
		{"\na", false, 1, "", nil},
		{"\na", true, 1, "", nil},
		{"\r\n", false, 2, "", nil},
		{"\r\n", true, 2, "", nil},
		{"\r\na", false, 2, "", nil},
		{"\r\na", true, 2, "", nil},
		{"\r", false, 0, "", nil}, // Waiting for \n
		{"\r", true, 1, "", nil},
		{"\ra", false, 1, "", nil},
		{"\ra", true, 1, "", nil},
		{"\n\r", false, 1, "", nil},
		{"\n\r", true, 1, "", nil},
		{"abc", false, 0, "", nil},
		{"abc", true, 3, "abc", nil},
		{"abc\n", false, 4, "abc", nil},
		{"abc\n", true, 4, "abc", nil},
		{"abc\nx", false, 4, "abc", nil},
		{"abc\nx", true, 4, "abc", nil},
		{"abc\n\n", false, 4, "abc", nil},
		{"abc\n\n", true, 4, "abc", nil},
		{"abc\r\n", false, 5, "abc", nil},
		{"abc\r\n", true, 5, "abc", nil},
		{"abc\r\nx", false, 5, "abc", nil},
		{"abc\r\nx", true, 5, "abc", nil},
		{"abc\r", false, 0, "", nil}, // Waiting for \n
		{"abc\r", true, 4, "abc", nil},
	} {
		t.Logf("in: %#v, atEOF: %v", test.in, test.atEOF)
		advance, line, err := scanLinesCR([]byte(test.in), test.atEOF)
		if test.err != nil {
			if assert.Error(err) {
				assert.Equal(0, advance)
				assert.Nil(line)
			}
		} else {
			if assert.NoError(err) {
				assert.Equal(test.advance, advance)
				assert.Equal(test.line, string(line))
			}
		}
	}
}
