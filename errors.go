package sse

import (
	"fmt"
	"net/http"
)

type (
	ErrHttpNotOk struct {
		code int
	}
	ErrContentType struct {
		contentType string
	}
)

func (me *ErrHttpNotOk) Error() string {
	return fmt.Sprintf("request status code was %d instead of %d", me.code, http.StatusOK)
}

func (me *ErrContentType) Error() string {
	return fmt.Sprintf("content type is %q instead of %q", me.contentType, ContentTypeSSE)
}
