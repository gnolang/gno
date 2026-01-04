package mux

type Handler struct {
	Pattern string
	Fn      HandlerFunc
}

type HandlerFunc func(*ResponseWriter, *Request)

type ErrHandlerFunc func(*ResponseWriter, *Request) error

type NotFoundHandler func(*ResponseWriter, *Request)

// TODO: AutomaticIndex
