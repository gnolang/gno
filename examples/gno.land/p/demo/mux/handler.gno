package mux

type HandlerFunc func(*ResponseWriter, *Request)

// TODO: ErrHandlerFunc...
// TODO: NotFoundHandler
// TODO: AutomaticIndex
