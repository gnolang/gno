package diff

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

const detailLimit = 8 << 10

var goroutineHeaderRE = regexp.MustCompile(`^goroutine \d+ \[`)

// RunGo compiles and runs src with the standard Go toolchain in a throwaway
// module. println output is collected from stderr, where the gc runtime
// writes the print and println builtins, and split off from any panic text.
func RunGo(ctx context.Context, src string, opt Options) Result {
	opt = opt.forGo()
	dir, err := os.MkdirTemp("", "gnodiff-go-*")
	if err != nil {
		return Result{Kind: KindInternal, Detail: err.Error()}
	}
	defer os.RemoveAll(dir)

	// Pin go 1.22 so per-iteration loop variables match Gno while newer
	// Go 1.23+ syntax does not slip past gc.
	mod := "module gnodiff.local/prog\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0o644); err != nil {
		return Result{Kind: KindInternal, Detail: err.Error()}
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0o644); err != nil {
		return Result{Kind: KindInternal, Detail: err.Error()}
	}

	runCtx, cancel := context.WithTimeout(ctx, opt.Timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, "go", "run", ".")
	cmd.Dir = dir
	configureCommand(cmd)
	var stdout, stderr cappedBuf
	stdout.init(opt.OutputLimit)
	stderr.init(opt.OutputLimit)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = goEnv()

	runErr := cmd.Run()
	timedOut := errors.Is(runCtx.Err(), context.DeadlineExceeded)
	return classifyGo(stdout.String(), stderr.String(), runErr, timedOut, stdout.truncated || stderr.truncated)
}

func goEnv() []string {
	return replaceEnv(os.Environ(), [][2]string{
		{"CGO_ENABLED", "0"},
		{"GO111MODULE", "on"},
		{"GOPROXY", "off"},
		{"GOSUMDB", "off"},
		{"GOTOOLCHAIN", "local"},
		{"GOWORK", "off"},
		{"GOFLAGS", "-mod=mod"},
	})
}

func replaceEnv(base []string, overrides [][2]string) []string {
	skip := make(map[string]bool, len(overrides))
	for _, kv := range overrides {
		skip[kv[0]] = true
	}
	out := make([]string, 0, len(base)+len(overrides))
	for _, entry := range base {
		key, _, ok := strings.Cut(entry, "=")
		if !ok || skip[key] {
			continue
		}
		out = append(out, entry)
	}
	for _, kv := range overrides {
		out = append(out, kv[0]+"="+kv[1])
	}
	return out
}

func combineOutput(stdout, stderr string) string {
	out := stdout
	if stderr != "" {
		if out != "" {
			out += "\n" + stderr
		} else {
			out = stderr
		}
	}
	return NormalizeOutput(out)
}

func classifyGo(stdout, stderr string, runErr error, timedOut, trunc bool) Result {
	if timedOut {
		return Result{
			Kind:      KindNonTerm,
			Stdout:    combineOutput(stdout, stderr),
			Detail:    "go run timed out",
			Truncated: trunc,
		}
	}
	if runErr == nil {
		return Result{
			Kind:      KindOK,
			Stdout:    combineOutput(stdout, stderr),
			Truncated: trunc,
		}
	}
	if signaled(runErr) {
		return Result{
			Kind:      KindNonTerm,
			Stdout:    combineOutput(stdout, stderr),
			Detail:    clip(runErr.Error(), detailLimit),
			Truncated: trunc,
		}
	}
	if prog, fatalMsg := splitGoFatal(stderr); fatalMsg != "" {
		return Result{
			Kind:      KindNonTerm,
			Stdout:    combineOutput(stdout, prog),
			Detail:    clip(fatalMsg, detailLimit),
			Truncated: trunc,
		}
	}
	if prog, panicMsg := splitGoStderr(stderr); panicMsg != "" {
		return Result{
			Kind:      KindPanic,
			Stdout:    combineOutput(stdout, prog),
			Panic:     panicMsg,
			Truncated: trunc,
		}
	}
	detail := strings.TrimSpace(stderr)
	if detail == "" {
		detail = runErr.Error()
	}
	return Result{
		Kind:      KindReject,
		Detail:    clip(detail, detailLimit),
		Truncated: trunc,
	}
}

func trimGoExitStatus(stderr string) []string {
	stderr = strings.ReplaceAll(stderr, "\r\n", "\n")
	lines := strings.Split(stderr, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "exit status ") {
		lines = lines[:len(lines)-1]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// splitGoFatal recognizes runtime fatal errors (stack overflow, out of memory)
// as non-termination rather than compile rejection.
func splitGoFatal(stderr string) (prog, fatalMsg string) {
	lines := trimGoExitStatus(stderr)
	for i := len(lines) - 1; i >= 0; i-- {
		after, ok := strings.CutPrefix(lines[i], "fatal error: ")
		if !ok {
			continue
		}
		if !hasStackHeader(lines[i+1:]) {
			continue
		}
		// Strip preceding Go runtime preamble lines (e.g. stack limit banner).
		cut := i
		for cut > 0 && strings.HasPrefix(lines[cut-1], "runtime: ") {
			cut--
		}
		return strings.Join(lines[:cut], "\n"), "fatal error: " + after
	}
	return "", ""
}

// splitGoStderr separates println output from a panic dump. The go command
// appends "exit status N" after an unrecovered panic. Scanning backwards for
// the last top-level "panic: " line before a goroutine stack header keeps
// earlier println("panic: ...") output intact and preserves the outermost
// message on nested defer panics (which are tab-indented).
func splitGoStderr(stderr string) (prog, panicMsg string) {
	lines := trimGoExitStatus(stderr)
	for i := len(lines) - 1; i >= 0; i-- {
		after, ok := strings.CutPrefix(lines[i], "panic: ")
		if !ok {
			continue
		}
		if hasGoroutineHeader(lines[i+1:]) {
			return strings.Join(lines[:i], "\n"), after
		}
	}
	return strings.Join(lines, "\n"), ""
}

func hasGoroutineHeader(lines []string) bool {
	return slices.ContainsFunc(lines, goroutineHeaderRE.MatchString)
}

func hasStackHeader(lines []string) bool {
	return slices.ContainsFunc(lines, func(ln string) bool {
		return ln == "runtime stack:" || goroutineHeaderRE.MatchString(ln)
	})
}

// cappedBuf collects output up to a ceiling and then discards the rest,
// reporting success to the writer so a chatty program is not failed for
// the cap itself.
type cappedBuf struct {
	buf       bytes.Buffer
	n         int
	max       int
	truncated bool
}

func (c *cappedBuf) init(limit int) {
	if limit < 1 {
		limit = 1
	}
	c.max = limit
}

func (c *cappedBuf) Write(p []byte) (int, error) {
	if c.max == 0 {
		c.max = defaultOutputLimit
	}
	if c.n >= c.max {
		c.truncated = true
		return len(p), nil
	}
	write := p
	if c.n+len(write) > c.max {
		write = write[:c.max-c.n]
		c.truncated = true
	}
	if _, err := c.buf.Write(write); err != nil {
		return 0, err
	}
	c.n += len(write)
	return len(p), nil
}

func (c *cappedBuf) String() string { return c.buf.String() }
