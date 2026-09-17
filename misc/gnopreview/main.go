// Command gnopreview builds a static gnoweb snapshot of what a pull request
// changed, so the PR can link to a rendered page instead of asking reviewers to
// run gnodev themselves.
//
// It has two modes, matching the two halves of the CI workflow:
//
//	gnopreview plan   -changed <file>          # what would be previewed, as JSON
//	gnopreview render -changed <file> -out dir # boot gnodev, crawl, write the tree
//
// `render` also writes <out>/preview.json and <out>/comment.md, which the
// publishing job turns into the sticky PR comment.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	defaultLive      = "https://gno.land"
	defaultMaxRealms = 25
	defaultMaxPages  = 400
)

type config struct {
	root      string
	baseRoot  string
	changed   string
	out       string
	gnodev    string
	port      int
	live      string
	baseURL   string
	pr        string
	maxRealms int
	maxPages  int
	chrome    string
	timeout   time.Duration
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "gnopreview:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: gnopreview <plan|render> [flags]")
	}
	cmd, args := args[0], args[1:]

	var cfg config
	fs := flag.NewFlagSet("gnopreview "+cmd, flag.ExitOnError)
	fs.StringVar(&cfg.root, "root", "", "gno monorepo root (default: walk up from the working directory)")
	fs.StringVar(&cfg.baseRoot, "base-root", "", "checkout of the merge base; enables before/after screenshots of changed realms")
	fs.StringVar(&cfg.changed, "changed", "-", "file holding the changed paths, one per line (- for stdin)")
	fs.StringVar(&cfg.out, "out", "_preview", "output directory")
	fs.StringVar(&cfg.gnodev, "gnodev", envOr("GNODEV", "gnodev"), "gnodev binary")
	fs.IntVar(&cfg.port, "port", 8899, "port gnodev serves gnoweb on")
	fs.StringVar(&cfg.live, "live", defaultLive, "origin used for links the snapshot does not contain")
	fs.StringVar(&cfg.baseURL, "base-url", "", "public URL the snapshot will be served from (for the comment)")
	fs.StringVar(&cfg.pr, "pr", "", "pull request number (for the comment)")
	fs.IntVar(&cfg.maxRealms, "max-realms", defaultMaxRealms, "cap on rendered realms; 0 for no cap")
	fs.IntVar(&cfg.maxPages, "max-pages", defaultMaxPages, "cap on crawled pages; 0 for no cap")
	fs.StringVar(&cfg.chrome, "chrome", "", "Chrome/Chromium binary for screenshots (default: autodetect)")
	fs.DurationVar(&cfg.timeout, "timeout", 5*time.Minute, "how long to wait for gnodev to come up")
	if err := fs.Parse(args); err != nil {
		return err
	}

	root, err := findRoot(cfg.root)
	if err != nil {
		return err
	}
	cfg.root = root

	changed, err := readLines(cfg.changed)
	if err != nil {
		return err
	}
	plan, err := BuildPlan(cfg.root, changed, cfg.maxRealms)
	if err != nil {
		return err
	}

	switch cmd {
	case "plan":
		return json.NewEncoder(os.Stdout).Encode(plan)
	case "render":
		return render(cfg, plan)
	}
	return fmt.Errorf("unknown command %q", cmd)
}

func render(cfg config, plan *Plan) error {
	if err := os.MkdirAll(cfg.out, 0o755); err != nil {
		return err
	}
	// An empty plan produces no comment.md at all: a PR that touches neither a
	// realm, a realm dependency nor gnoweb gets no comment, not an empty one.
	if plan.Empty() {
		fmt.Println("nothing to preview")
		return writeJSON(filepath.Join(cfg.out, "preview.json"), plan)
	}

	fmt.Printf("rendering %d realm(s): %s\n", len(plan.Realms), strings.Join(plan.Realms, " "))
	stop, err := startGnodev(cfg, cfg.root, plan.Dirs, cfg.port, "gnodev.log")
	if err != nil {
		return err
	}
	defer stop()

	c := &Crawler{
		Base:     fmt.Sprintf("http://127.0.0.1:%d", cfg.port),
		Realms:   plan.Realms,
		MaxPages: cfg.maxPages,
		Live:     strings.TrimSuffix(cfg.live, "/"),
	}
	if err := waitReady(c.Base, urlOf(plan.Realms[0]), cfg.timeout); err != nil {
		return err
	}
	if err := c.Run(); err != nil {
		return err
	}
	assets := filepath.Join(cfg.root, "gno.land", "pkg", "gnoweb", "public")
	if err := c.Write(cfg.out, assets); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(cfg.out, "index.html"), Index(plan, c)); err != nil {
		return err
	}
	// Screenshots serve two different questions. When gnoweb itself changed,
	// a fixed sample of pages shows what the chrome now looks like. When a
	// realm changed, the useful picture is that realm before and after.
	if len(plan.ChangedRealms) > 0 {
		base := renderBase(cfg, plan, c)
		plan.Pairs = ScreenshotPairs(cfg.out, c, base, plan.ChangedRealms, cfg.chrome)
	} else if plan.Gnoweb {
		plan.Shots = Screenshot(cfg.out, c, cfg.chrome)
	}
	if err := writeJSON(filepath.Join(cfg.out, "preview.json"), plan); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(cfg.out, "comment.md"), Comment(plan, cfg.baseURL, cfg.pr)); err != nil {
		return err
	}
	fmt.Printf("done: %d page(s), %d screenshot(s), %d before/after pair(s) -> %s\n",
		len(c.pages), len(plan.Shots), len(plan.Pairs), cfg.out)
	return nil
}

// startGnodev boots gnodev on the given package dirs of the given tree.
// Dependencies resolve lazily out of examples/, so only the realms being
// previewed are loaded.
func startGnodev(cfg config, root string, dirs []string, port int, logName string) (func(), error) {
	examples := filepath.Join(root, examplesRel)
	args := []string{
		"local", "-no-watch",
		"-web-listener", fmt.Sprintf("127.0.0.1:%d", port),
		// The RPC listener and the keybase both default to fixed locations
		// (127.0.0.1:26657 and $GNOHOME), so the before/after passes — which
		// run at the same time — would collide on them. Derive both from the
		// web port instead.
		"-node-rpc-listener", fmt.Sprintf("tcp://127.0.0.1:%d", port+10000),
		"-home", filepath.Join(cfg.out, fmt.Sprintf(".gnodev-%d", port)),
		"-C", examples,
	}
	for _, d := range dirs {
		args = append(args, filepath.Join(root, filepath.FromSlash(d)))
	}
	log, err := os.Create(filepath.Join(cfg.out, logName))
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(cfg.gnodev, args...)
	cmd.Env = append(os.Environ(), "GNOROOT="+root)
	cmd.Stdout, cmd.Stderr = log, log
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		log.Close()
		return nil, fmt.Errorf("start gnodev: %w", err)
	}
	return func() {
		// gnodev spawns a node; kill the whole process group.
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
		_ = cmd.Wait()
		log.Close()
	}, nil
}

// renderBase renders the changed realms a second time from the merge-base
// checkout, into <out>/_before/. It reuses the head's gnodev binary and the
// head's assets on purpose: the pair must differ by the realm change alone, not
// by whatever else moved on master. Returns nil when there is no base checkout,
// none of the changed realms exist there (all new), or the pass fails — a
// missing "before" costs the comment one image, not the preview.
func renderBase(cfg config, plan *Plan, head *Crawler) *Crawler {
	if cfg.baseRoot == "" {
		return nil
	}
	var realms, dirs []string
	for i, r := range plan.Realms {
		if !contains(plan.ChangedRealms, r) {
			continue
		}
		if _, err := os.Stat(filepath.Join(cfg.baseRoot, filepath.FromSlash(plan.Dirs[i]))); err != nil {
			continue // added by this pull request
		}
		realms = append(realms, r)
		dirs = append(dirs, plan.Dirs[i])
	}
	if len(realms) == 0 {
		return nil
	}
	port := cfg.port + 1
	stop, err := startGnodev(cfg, cfg.baseRoot, dirs, port, "gnodev-base.log")
	if err != nil {
		fmt.Fprintln(os.Stderr, "  ! base render:", err)
		return nil
	}
	defer stop()

	base := &Crawler{
		Base:       fmt.Sprintf("http://127.0.0.1:%d", port),
		Realms:     realms,
		MaxPages:   len(realms),
		Live:       head.Live,
		RenderOnly: true,
		Prefix:     beforeDir,
	}
	if err := waitReady(base.Base, urlOf(realms[0]), cfg.timeout); err != nil {
		fmt.Fprintln(os.Stderr, "  ! base render:", err)
		return nil
	}
	if err := base.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "  ! base render:", err)
		return nil
	}
	if err := base.Write(cfg.out, ""); err != nil {
		fmt.Fprintln(os.Stderr, "  ! base render:", err)
		return nil
	}
	return base
}

// --- small helpers ---------------------------------------------------------

// findRoot walks up from dir (or the working directory) until it finds the
// monorepo root, identified by examples/gnowork.toml.
func findRoot(dir string) (string, error) {
	if dir != "" {
		return filepath.Abs(dir)
	}
	cur, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(cur, examplesRel, "gnowork.toml")); err == nil {
			return cur, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", errors.New("gno monorepo root not found (no examples/gnowork.toml above the working directory); pass -root")
		}
		cur = parent
	}
}

func readLines(p string) ([]string, error) {
	var b []byte
	var err error
	if p == "-" {
		b, err = os.ReadFile("/dev/stdin")
	} else {
		b, err = os.ReadFile(p)
	}
	if err != nil {
		return nil, err
	}
	var out []string
	for l := range strings.SplitSeq(string(b), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out, nil
}

func writeJSON(p string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(p, string(b)+"\n")
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
