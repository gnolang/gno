package main

import (
	"embed"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/contribs/gnobro/pkg/browser"
)

//go:embed assets/*.utf8ans
var gnoland_banner embed.FS

func NewGnoLandBanner() browser.ModelBanner {
	const assets = "assets"

	entries, err := gnoland_banner.ReadDir(assets)
	if err != nil {
		panic("unable to banner dir: " + err.Error())
	}

	frames := make([]string, len(entries))
	for i, entry := range entries {
		if entry.IsDir() {
			continue
		}

		frame, err := gnoland_banner.ReadFile(filepath.Join(assets, entry.Name()))
		if err != nil {
			panic("unable to read banner frame: " + err.Error())
		}

		frames[i] = string(frame)
	}

	return browser.NewModelBanner(time.Second/3, frames)
}
