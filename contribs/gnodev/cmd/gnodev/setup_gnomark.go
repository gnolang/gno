// File: contribs/gnodev/cmd/gnodev/setup_gnomark.go
//go:build gnomark

package main

import (
	"fmt"
	"github.com/yuin/goldmark"
	"log/slog"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	mdhtml "github.com/yuin/goldmark/renderer/html"
)

func init() {
	gnoweb.SetRenderFactory(RenderWithGnoMark)
	fmt.Printf("Added GnoMark renderer to gnoweb\n")
}

// RenderWithGnoMark creates a MarkdownRenderer configured with GnoMark support.
func RenderWithGnoMark(logger *slog.Logger, cfg *gnoweb.AppConfig) (*gnoweb.MarkdownRenderer, error) {
	markdownCfg := gnoweb.NewDefaultMarkdownRendererConfig(nil)
	if cfg.UnsafeHTML {
		markdownCfg.GoldmarkOptions = append(markdownCfg.GoldmarkOptions, goldmark.WithRendererOptions(
			mdhtml.WithXHTML(), mdhtml.WithUnsafe(),
		))
	}
	md := goldmark.New(markdownCfg.GoldmarkOptions...)
	// NOTE: users extending gnoweb can load their own extensions here.
	// CustomExtension().Extend(md)
	return &gnoweb.MarkdownRenderer{
		Logger:   logger,
		Markdown: md,
	}, nil
}
