package mux

type HandlerFunc func(*ResponseWriter, *Request)

// TODO: type ErrHandlerFunc func(*ResponseWriter, *Request) error
// TODO: NotFoundHandler
// TODO: AutomaticIndex
