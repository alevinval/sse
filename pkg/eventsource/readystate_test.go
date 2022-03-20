package eventsource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadyState_EnumValuesMatch(t *testing.T) {
	for _, test := range []struct {
		number   byte
		expected ReadyState
	}{
		{0, Connecting},
		{1, Open},
		{2, Closed},
	} {
		assert.Equal(t, test.expected, ReadyState(test.number))
	}
}
