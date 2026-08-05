package mux

// Handler stores a route pattern with one of two handler shapes.
// Fn (HandlerFunc, no rlm) is set by HandleFunc; FnRlm (HandlerFuncRlm,
// rlm-aware non-crossing) is set by HandleFuncRlm. Exactly one is set
// per route. RenderRlm dispatches FnRlm with the supplied rlm; Render
// dispatches Fn and panics if the matched route was registered with
// HandleFuncRlm (caller used the wrong dispatch method).
type Handler struct {
	Pattern string
	Fn      HandlerFunc
	FnRlm   HandlerFuncRlm
}

type HandlerFunc func(*ResponseWriter, *Request)

// HandlerFuncRlm is the rlm-aware handler shape — non-crossing, realm
// last, so callers thread cur as data for the handler to forward to
// downstream crossing functions.
type HandlerFuncRlm func(res *ResponseWriter, req *Request, rlm realm)

type ErrHandlerFunc func(*ResponseWriter, *Request) error

type NotFoundHandler func(*ResponseWriter, *Request)

// TODO: AutomaticIndex
