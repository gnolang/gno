package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/contribs/gnodev/pkg/browser"
)

//go:embed assets/*.utf8ans
var banner_gnoland embed.FS

func NewBanner_GnoLand() browser.ModelBanner {
	const assets = "assets"

	entries, err := banner_gnoland.ReadDir(assets)
	if err != nil {
		panic("unable to banner dir: " + err.Error())
	}

	frames := make([]string, len(entries))
	for i, entry := range entries {
		if entry.IsDir() {
			continue
		}

		frame, err := banner_gnoland.ReadFile(filepath.Join(assets, entry.Name()))
		if err != nil {
			panic("unable to read banner frame: " + err.Error())
		}

		os.Stdout.Write(frame)
		fmt.Println()
		frames[i] = string(frame)
	}

	return browser.NewModelBanner(time.Second/3, frames)
}
