package sse

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithBasicAuth(t *testing.T) {
	r := &http.Request{Header: http.Header{}}

	WithBasicAuth("user", "password")(r)

	assert.Equal(t, r.Header.Get("Authorization"), "Basic dXNlcjpwYXNzd29yZA==")
}

func TestWithBearerTokenAuth(t *testing.T) {
	r := &http.Request{Header: http.Header{}}

	WithBearerTokenAuth("token-value")(r)

	assert.Equal(t, r.Header.Get("Authorization"), "Bearer token-value")
}
