package sse_test

import (
	"github.com/mubit/sse"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestProperEventInitialisation(t *testing.T) {
	test_data := []byte("test event data")
	event := sse.NewEvent("test_id", "test_name", test_data)

	assert.Equal(t, "test_id", event.Id())
	assert.Equal(t, "test_name", event.Name())
	assert.Equal(t, test_data, event.Data())
}

func TestInitialiseEventCopiesDataBuffer(t *testing.T) {
	test_data := []byte("test event data")
	event := sse.NewEvent("test_id", "test_name", test_data)

	// Pointers should differ, because NewEvent() allocates a new buffer
	// and copies data over.
	p1 := reflect.ValueOf(test_data).Pointer()
	p2 := reflect.ValueOf(event.Data()).Pointer()

	assert.NotEqual(t, p1, p2)
}
