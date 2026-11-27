package diff

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
)

const detailLimit = 8 << 10

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

	mod := "module gnodiff.local/prog\n\ngo 1.25\n"
	if err := os.WriteFile(dir+"/go.mod", []byte(mod), 0o644); err != nil {
		return Result{Kind: KindInternal, Detail: err.Error()}
	}
	if err := os.WriteFile(dir+"/main.go", []byte(src), 0o644); err != nil {
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
	return replaceEnv(os.Environ(), map[string]string{
		"CGO_ENABLED": "0",
		"GO111MODULE": "on",
		"GOPROXY":     "off",
		"GOSUMDB":     "off",
		"GOTOOLCHAIN": "local",
		"GOFLAGS":     "-mod=mod",
	})
}

func replaceEnv(base []string, kv map[string]string) []string {
	out := make([]string, 0, len(base)+len(kv))
	for _, entry := range base {
		key, _, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		if _, replace := kv[key]; replace {
			continue
		}
		out = append(out, entry)
	}
	for key, val := range kv {
		out = append(out, key+"="+val)
	}
	return out
}

func classifyGo(stdout, stderr string, runErr error, timedOut, trunc bool) Result {
	prog, panicMsg := splitGoStderr(stderr)
	out := stdout
	if prog != "" {
		if out != "" {
			out += "\n" + prog
		} else {
			out = prog
		}
	}
	res := Result{Stdout: NormalizeOutput(out), Truncated: trunc}
	if timedOut {
		res.Kind = KindNonTerm
		res.Detail = "go run timed out"
		return res
	}
	if panicMsg != "" {
		res.Kind = KindPanic
		res.Panic = panicMsg
		return res
	}
	if runErr == nil {
		res.Kind = KindOK
		return res
	}
	if signaled(runErr) {
		res.Kind = KindNonTerm
		res.Detail = clip(runErr.Error(), detailLimit)
		return res
	}
	res.Kind = KindReject
	res.Stdout = ""
	detail := strings.TrimSpace(stderr)
	if detail == "" {
		detail = runErr.Error()
	}
	res.Detail = clip(detail, detailLimit)
	return res
}

// splitGoStderr separates println output from a panic dump. The go command
// appends "exit status N" after an unrecovered panic. A line that merely
// starts with "panic: " (a println of that prefix) is program output unless
// a goroutine stack follows it.
func splitGoStderr(stderr string) (prog, panicMsg string) {
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
	for i, line := range lines {
		after, ok := strings.CutPrefix(line, "panic: ")
		if !ok {
			continue
		}
		if strings.Contains(strings.Join(lines[i+1:], "\n"), "goroutine ") {
			return strings.Join(lines[:i], "\n"), after
		}
	}
	return strings.Join(lines, "\n"), ""
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
		c.max = 256 << 10
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
