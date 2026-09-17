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
