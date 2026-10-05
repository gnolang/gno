// Package store renders Explore, the app store realm (gno.land/adr/adr-004-gnoweb-store.md).
//
// The store is a realm like any other: gnoweb shows it at its own path, with
// its breadcrumb, tabs and omnibar. This package only replaces the markdown
// of its Content tab with a richer view built from the realm's api/v1 JSON.
// Every field is untrusted; what a listing may show is decided here, from
// the operator's trust list. It never writes to the chain and keeps no state
// beyond a short-lived response cache.
package store

import (
	"context"
	"html/template"
	"log/slog"
)

// ClientAdapter is the subset of gnoweb.ClientAdapter the store consumes,
// declared locally because gnoweb imports this package.
type ClientAdapter interface {
	Realm(ctx context.Context, path, args string) ([]byte, error)
}

// Deps is a struct of interfaces so each field is independently mockable.
type Deps struct {
	Client ClientAdapter

	// RealmPath is the registry realm as a gnoweb path ("/r/gnoland/store/v0").
	RealmPath string

	// Domain prefixes on-chain package paths ("gno.land").
	Domain string

	// Trusted reports whether a package path, without domain and "/r/" or
	// "/p/" ("gnoland/blog"), is under the operator's trusted list. It is
	// the same predicate that drives the community-realm notice.
	Trusted func(pkg string) bool

	Logger *slog.Logger
}

// Handler renders the store realm's views.
type Handler struct {
	deps      Deps
	cache     *responseCache
	templates *template.Template
}

// New validates required deps and returns a Handler.
// Panics on a missing dependency; Logger falls back to slog.Default().
func New(deps Deps) *Handler {
	switch {
	case deps.Client == nil:
		panic("store.New: Client is required")
	case deps.Trusted == nil:
		panic("store.New: Trusted is required")
	case deps.RealmPath == "" || deps.Domain == "":
		panic("store.New: RealmPath and Domain are required")
	}
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}
	return &Handler{deps: deps, cache: newResponseCache(), templates: newTemplates(deps.RealmPath, deps.Domain)}
}

// RealmPath is the path the store owns; gnoweb links to it from the header.
func (h *Handler) RealmPath() string { return h.deps.RealmPath }
