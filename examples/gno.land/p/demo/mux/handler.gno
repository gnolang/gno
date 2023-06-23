package mux

type Handler struct {
	Pattern string
	Fn      HandlerFunc
}

type HandlerFunc func(*ResponseWriter, *Request)

// TODO: type ErrHandlerFunc func(*ResponseWriter, *Request) error
// TODO: NotFoundHandler
// TODO: AutomaticIndex
