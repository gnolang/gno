package httptest

import (
	"bytes"

	"filetests/extern/net/http"
)

type ResponseRecorder struct {
	// Code is the HTTP response code set by WriteHeader.
	//
	// Note that if a Handler never calls WriteHeader or Write,
	// this might end up being 0, rather than the implicit
	// http.StatusOK. To get the implicit value, use the Result
	// method.
	Code int

	// HeaderMap contains the headers explicitly set by the Handler.
	// It is an internal detail.
	//
	// Deprecated: HeaderMap exists for historical compatibility
	// and should not be used. To access the headers returned by a handler,
	// use the Response.Header map as returned by the Result method.
	HeaderMap http.Header

	// Body is the buffer to which the Handler's Write calls are sent.
	// If nil, the Writes are silently discarded.
	Body *bytes.Buffer

	// Flushed is whether the Handler called Flush.
	Flushed bool
	// contains filtered or unexported fields
}

func NewRecorder() *ResponseRecorder {
	return &ResponseRecorder{
		HeaderMap: make(http.Header),
		Body:      new(bytes.Buffer),
		Code:      200,
	}
}

// XXX dummy
func (rw *ResponseRecorder) Header() http.Header {
	return rw.HeaderMap
}

// XXX dummy
func (rw *ResponseRecorder) Write(buf []byte) (int, error) {
	return 0, nil
}

// XXX dummy
func (rw *ResponseRecorder) WriteHeader(code int) {
}
