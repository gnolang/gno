package mux

import "strings"

// ResponseWriter represents the response writer.
type ResponseWriter struct {
	output strings.Builder
}

// Write appends data to the response output.
func (rw *ResponseWriter) Write(data string) {
	rw.output.WriteString(data)
}

// Output returns the final response output.
func (rw *ResponseWriter) Output() string {
	return rw.output.String()
}

// TODO: func (rw *ResponseWriter) Header()...
