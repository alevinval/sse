package sse

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProperEventInitialisation(t *testing.T) {
	data := []byte("test event data")
	event := newEvent("test_id", "test_name", data)

	assert.Equal(t, "test_id", event.ID())
	assert.Equal(t, "test_name", event.Name())
	assert.Equal(t, data, event.Data())
}

func TestInitialiseEventCopiesDataBuffer(t *testing.T) {
	data := []byte("test event data")
	event := newEvent("test_id", "test_name", data)

	// Equal doesn't check pointers but values
	assert.Equal(t, data, event.Data())

	// Pointers should differ, because NewEvent() allocates a new buffer
	// and copies data over.
	p1 := reflect.ValueOf(data).Pointer()
	p2 := reflect.ValueOf(event.Data()).Pointer()

	assert.NotEqual(t, p1, p2)
}
