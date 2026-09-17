package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// Shot is one screenshot embedded in the PR comment.
type Shot struct {
	File  string `json:"file"`  // path inside the snapshot, e.g. _shots/home.png
	Label string `json:"label"` // caption
}

// shotsDir is where screenshots land inside the published snapshot, so the PR
// comment can embed them by URL without uploading anything.
const shotsDir = "_shots"

// beforeDir holds the same realms rendered from the merge base.
const beforeDir = "_before"

// maxPairs caps how many changed realms get a before/after pair. Four images is
// already a lot of comment; the full list of realms is right underneath.
const maxPairs = 2

// ShotPair is one realm shown before and after the pull request's change.
type ShotPair struct {
	Realm  string `json:"realm"`
	Before string `json:"before,omitempty"` // empty when the realm is new in this PR
	After  string `json:"after"`
	URL    string `json:"url"` // the after page, for the link behind the image
}

// ScreenshotPairs photographs each changed realm as the merge base renders it
// and as this branch renders it. Both passes use the SAME gnoweb — the binary
// and the assets come from the head — so what the pair shows is the realm
// change and nothing else.
func ScreenshotPairs(outDir string, head, base *Crawler, realms []string, chrome string) []ShotPair {
	bin := findChrome(chrome)
	if bin == "" {
		fmt.Fprintln(os.Stderr, "  ! no Chrome/Chromium found — publishing the preview without before/after")
		return nil
	}
	if err := os.MkdirAll(filepath.Join(outDir, shotsDir), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "  !", err)
		return nil
	}
	var pairs []ShotPair
	for _, r := range realms {
		if len(pairs) >= maxPairs {
			fmt.Fprintf(os.Stderr, "  i before/after capped at %d realm(s); the rest are listed as links\n", maxPairs)
			break
		}
		afterFile, ok := head.FileOf(urlOf(r))
		if !ok {
			continue
		}
		name := slug(strings.TrimPrefix(urlOf(r), "/"))
		pair := ShotPair{Realm: r, URL: path.Dir(afterFile) + "/"}
		if err := chromeShot(bin, filepath.Join(outDir, filepath.FromSlash(afterFile)),
			filepath.Join(outDir, shotsDir, name+"-after.png")); err != nil {
			fmt.Fprintf(os.Stderr, "  ! screenshot %s (after): %v\n", r, err)
			continue
		}
		pair.After = path.Join(shotsDir, name+"-after.png")

		// A realm added by this pull request has no "before"; say so rather
		// than inventing one.
		if base != nil {
			if beforeFile, ok := base.FileOf(urlOf(r)); ok {
				if err := chromeShot(bin, filepath.Join(outDir, filepath.FromSlash(beforeFile)),
					filepath.Join(outDir, shotsDir, name+"-before.png")); err == nil {
					pair.Before = path.Join(shotsDir, name+"-before.png")
				} else {
					fmt.Fprintf(os.Stderr, "  ! screenshot %s (before): %v\n", r, err)
				}
			}
		}
		pairs = append(pairs, pair)
		fmt.Printf("  📷 %s (before/after)\n", r)
	}
	return pairs
}

// shotPlan is the fixed sample photographed when gnoweb itself changed: the
// four views that carry most of gnoweb's chrome. Each entry is tried in order
// and skipped when the page was not captured.
var shotPlan = []struct{ url, label, name string }{
	{"/r/gnoland/home", "Home — rendered markdown", "home"},
	{"/r/gnoland/home$source", "Source view", "source"},
	{"/r/gnoland/boards2/v0$help", "Help / actions forms", "help"},
	{"/r/docs/security_patterns", "Long-form docs", "docs"},
}

// Screenshot renders pages of the finished snapshot with headless Chrome. It is
// best-effort: a missing browser costs the comment its images, not the preview.
func Screenshot(outDir string, c *Crawler, chrome string) []Shot {
	bin := findChrome(chrome)
	if bin == "" {
		fmt.Fprintln(os.Stderr, "  ! no Chrome/Chromium found — publishing the preview without screenshots")
		return nil
	}
	if err := os.MkdirAll(filepath.Join(outDir, shotsDir), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "  !", err)
		return nil
	}
	var shots []Shot
	for _, s := range shotPlan {
		p, ok := c.pages[canonicalURL(s.url)]
		if !ok {
			continue
		}
		src := filepath.Join(outDir, filepath.FromSlash(p.File))
		dst := filepath.Join(outDir, shotsDir, s.name+".png")
		if err := chromeShot(bin, src, dst); err != nil {
			fmt.Fprintf(os.Stderr, "  ! screenshot %s: %v\n", s.url, err)
			continue
		}
		shots = append(shots, Shot{File: path.Join(shotsDir, s.name+".png"), Label: s.label})
		fmt.Printf("  📷 %s\n", s.url)
	}
	return shots
}

func chromeShot(bin, src, dst string) error {
	abs, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	out, err := filepath.Abs(dst)
	if err != nil {
		return err
	}
	cmd := exec.Command(bin,
		"--headless=new",
		"--disable-gpu",
		"--no-sandbox",
		"--hide-scrollbars",
		"--force-device-scale-factor=1",
		"--window-size=1280,860",
		// Let the controller modules load and the webfonts settle before the
		// frame is grabbed; without it the shot is unstyled text.
		"--virtual-time-budget=4000",
		"--screenshot="+out,
		"file://"+abs,
	)
	if b, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(b)))
	}
	if fi, err := os.Stat(out); err != nil || fi.Size() == 0 {
		return fmt.Errorf("chrome wrote no image")
	}
	return nil
}

// findChrome resolves a browser binary: the explicit flag, then $CHROME, then
// the usual Linux names, then the macOS bundle path.
func findChrome(explicit string) string {
	candidates := []string{explicit, os.Getenv("CHROME")}
	candidates = append(candidates,
		"google-chrome", "google-chrome-stable", "chromium", "chromium-browser",
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
	)
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if p, err := exec.LookPath(c); err == nil {
			return p
		}
		if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
			return c
		}
	}
	return ""
}
