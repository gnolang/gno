package main

import (
	"context"
	"flag"
	"net/url"
	"os/exec"
	"runtime"
	"runtime/debug"
	"strings"
	"text/template"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

// NOTE: keep in sync with .github/ISSUE_TEMPLATE/BUG-REPORT.md
const bugTmpl = `## [Subject of the issue]

### Description

Describe your issue in as much detail as possible here

### Your environment

* go version {{.GoVersion}} {{.Os}}/{{.Arch}}
* gno commit that causes this issue: {{.Commit}}

### Steps to reproduce

* Tell us how to reproduce this issue
* Where the issue is, if you know
* Which commands triggered the issue, if any

### Expected behaviour

Tell us what should happen

### Actual behaviour

Tell us what happens instead

### Logs

Please paste any logs here that demonstrate the issue, if they exist

### Proposed solution

If you have an idea of how to fix this issue, please write it down here, so we can begin discussing it

`

type bugCfg struct {
	skipBrowser bool
}

func newBugCmd(io commands.IO) *commands.Command {
	cfg := &bugCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "bug",
			ShortUsage: "bug",
			ShortHelp:  "Start a bug report",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execBug(cfg, args, io)
		},
	)
}

func (c *bugCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.skipBrowser,
		"skip-browser",
		false,
		"do not open the browser",
	)
}

func execBug(cfg *bugCfg, args []string, io commands.IO) error {
	if len(args) != 0 {
		return flag.ErrHelp
	}

	bugReportEnv := struct {
		Os, Arch, GoVersion, Commit string
	}{
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version(),
		getCommitHash(),
	}

	var buf strings.Builder
	tmpl, err := template.New("bug.tmpl").Parse(bugTmpl)
	if err != nil {
		return err
	}
	tmpl.Execute(&buf, bugReportEnv)

	body := buf.String()
	url := "https://github.com/gnolang/gno/issues/new?body=" + url.QueryEscape(body)

	if !cfg.skipBrowser && openBrowser(url) {
		return nil
	}

	io.Println("Please file a new issue at github.com/gnolang/gno/issues/new using this template:")
	io.Println()
	io.Println(body)

	return nil
}

// openBrowser opens a default web browser with the specified URL.
func openBrowser(url string) bool {
	var cmdArgs []string
	switch runtime.GOOS {
	case "windows":
		cmdArgs = []string{"cmd", "/c", "start", url}
	case "darwin":
		cmdArgs = []string{"/usr/bin/open", url}
	default: // "linux"
		cmdArgs = []string{"xdg-open", url}
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	if cmd.Start() == nil && appearsSuccessful(cmd, 3*time.Second) {
		return true
	}

	return false
}

// getCommitHash returns the commit hash from build info, or an
// empty string if not found.
func getCommitHash() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return ""
}

// appearsSuccessful reports whether the command appears to have run successfully.
// If the command runs longer than the timeout, it's deemed successful.
// If the command runs within the timeout, it's deemed successful if it exited cleanly.
// Note: Taken from Go's `internal/browser“
func appearsSuccessful(cmd *exec.Cmd, timeout time.Duration) bool {
	errc := make(chan error, 1)
	go func() {
		errc <- cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		return true
	case err := <-errc:
		return err == nil
	}
}
