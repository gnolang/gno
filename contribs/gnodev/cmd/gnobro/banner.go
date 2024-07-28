package main

import (
	_ "embed"

	"strings"
	"time"

	"github.com/gnolang/gno/contribs/gnodev/pkg/browser"
)

//go:embed assets/gnoland-ansi-pink.utf8ans
var banner_gnoland string

func NewBanner_GnoLand() browser.ModelBanner {
	r := strings.NewReader(banner_gnoland)
	return browser.NewModelBanner(time.Second/50, r)
}
