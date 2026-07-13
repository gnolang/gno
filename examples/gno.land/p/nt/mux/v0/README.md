> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `mux` - Path router for Render

Simple routing and rendering library for `Render(path)` requests in Gno realms. Similar in spirit to `http.ServeMux`, with support for path variables (`{name}`), wildcards (`*`), and query strings.

## Usage

```go
package myrealm

import "gno.land/p/nt/mux/v0"

var router *mux.Router

func init() {
    router = mux.NewRouter()

    // Static route.
    router.HandleFunc("", func(res *mux.ResponseWriter, req *mux.Request) {
        res.Write("# Home\n")
    })

    // Named parameter.
    router.HandleFunc("hello/{name}", func(res *mux.ResponseWriter, req *mux.Request) {
        name := req.GetVar("name")
        res.Write("Hello, " + name + "!")
    })

    // Query string.
    router.HandleFunc("search", func(res *mux.ResponseWriter, req *mux.Request) {
        q := req.Query.Get("q")
        res.Write("Searching for: " + q)
    })

    // Wildcard - matches the rest of the path.
    router.HandleFunc("files/*", func(res *mux.ResponseWriter, req *mux.Request) {
        res.Write("File path: " + req.GetVar("*"))
    })
}

// Realm entry point.
func Render(path string) string {
    return router.Render(path)
}
```

## API

```go
type Router struct {
    NotFoundHandler NotFoundHandler
    // unexported
}

func NewRouter() *Router

func (r *Router) HandleFunc(pattern string, fn HandlerFunc)
func (r *Router) HandleFuncRlm(pattern string, fn HandlerFuncRlm) // rlm-aware handler
func (r *Router) HandleErrFunc(pattern string, fn ErrHandlerFunc)
func (r *Router) SetNotFoundHandler(handler NotFoundHandler)
func (r *Router) Render(reqPath string) string
func (r *Router) RenderRlm(_ int, rlm realm, reqPath string) string // dispatches rlm-aware routes

type Request struct {
    Path        string     // path without query string
    RawPath     string     // path including "?..." query string
    HandlerPath string     // pattern that matched this request
    Query       url.Values // parsed query parameters
}

func (r *Request) GetVar(key string) string

type ResponseWriter struct{ /* unexported */ }

func (rw *ResponseWriter) Write(data string)
func (rw *ResponseWriter) Output() string

type Handler struct {
    Pattern string
    Fn      HandlerFunc    // set by HandleFunc
    FnRlm   HandlerFuncRlm // set by HandleFuncRlm
}

type HandlerFunc     func(*ResponseWriter, *Request)
type HandlerFuncRlm  func(_ int, rlm realm, res *ResponseWriter, req *Request)
type ErrHandlerFunc  func(*ResponseWriter, *Request) error
type NotFoundHandler func(*ResponseWriter, *Request)
```

## Route patterns

- `users` - static, matches exactly `users`.
- `users/{id}` - named parameter, extracted with `req.GetVar("id")`.
- `files/*` - wildcard, captures all remaining segments. Extract with `req.GetVar("*")`.

Routes are matched in registration order; the first match wins. If no route matches, `NotFoundHandler` runs (default writes `"404"`).

## Notes

- `HandleErrFunc` wraps an error-returning handler: a non-nil error is written as `"Error: " + err.Error()` to the response.
- Query strings are parsed off `reqPath` (`?foo=bar`); access via `req.Query` (a `net/url.Values`).
- `req.RawPath` keeps the original path including the query string; `req.Path` strips it.
- `req.GetVar(...)` and `req.Query.Get(...)` return attacker-controlled path/query input. Wrap it with `sanitize.InlineText` from [`gno.land/p/nt/markdown/sanitize/v0`](../../markdown/sanitize/v0) before writing it into the response, or user input can inject Markdown structure.
- Register realm-aware handlers with `HandleFuncRlm` and dispatch them with `RenderRlm(0, cur, path)`. The plain `Render` path only invokes non-rlm `Fn` handlers.
