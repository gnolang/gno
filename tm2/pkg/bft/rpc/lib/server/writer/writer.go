package writer

// ResponseWriter outlines the interface any
// JSON-RPC response writer needs to implement
type ResponseWriter interface {
	// WriteResponse takes in the JSON-RPC response
	// which is either a single object, or a batch
	WriteResponse(response any)
}
