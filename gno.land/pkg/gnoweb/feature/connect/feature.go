package connect

import (
	"log/slog"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

// PageMeta is the subset of gnoweb's StaticMetadata the install page needs.
// Declared locally so feature/connect does not import the gnoweb package
// (app.go imports connect — a back import would create a cycle).
type PageMeta struct {
	AssetsPath        string
	ChromaPath        string
	ChainId           string
	Remote            string
	BuildTime         string
	AnalyticsHostname string
	Analytics         bool
	Banner            components.BannerData
}

// Deps is a struct of fields so each is independently settable in tests.
type Deps struct {
	Meta   PageMeta
	Logger *slog.Logger
}

type Handler struct {
	deps Deps
}

// New returns a Handler. Logger falls back to slog.Default().
func New(deps Deps) *Handler {
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}
	return &Handler{deps: deps}
}
