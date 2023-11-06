package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func newBugCmd(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "bug",
			ShortUsage: "bug",
			ShortHelp:  "start a bug report",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execBug(args, io)
		},
	)
}

func execBug(args []string, io commands.IO) error {
	if len(args) != 0 {
		return flag.ErrHelp
	}

	var buf strings.Builder
	buf.WriteString(bugHeader)
	writeEnvironment(&buf)
	buf.WriteString(bugFooter)
	// TODO: include gno version or commit?

	body := buf.String()
	url := "https://github.com/gnolang/gno/issues/new?body=" + url.QueryEscape(body)

	if !openBrowser(url) {
		io.Println("Please file a new issue at github.com/gnolang/gno/issues/new using this template:")
		io.Println()
		io.Println(body)
	}

	return nil
}

const bugHeader = `<!-- Please answer these questions before submitting your issue. Thanks! -->

`

const bugFooter = `### What did you do?

<!--
If possible, provide a recipe for reproducing the error.
-->

### What did you expect to see?



### What did you see instead?

`

// openBrowser opens a default web browser with the specified URL.
// return true if it started successfully within a timeout (3 Second).
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
	if cmd.Start() == nil && isExecutionSuccessfulWithinTimeout(cmd, 3*time.Second) {
		return true
	}
	return false
}

// writeEnvironment writes environment information to the provided io.Writer.
// It includes Go version details and OS details within code blocks.
func writeEnvironment(w io.Writer) {
	fmt.Fprintf(w, "### Environment\n\n")
	fmt.Fprintf(w, "```\n")
	writeGoVersion(w)
	writeOSDetails(w)
	fmt.Fprintf(w, "```\n\n")
}

// writeGoVersion writes Go version information to the io.Writer.
func writeGoVersion(w io.Writer) {
	fmt.Fprintf(w, "$ go version\n")
	fmt.Fprintf(w, "go version %s %s/%s\n\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

// writeOSDetails writes OS details (uses `uname`) to the io.Writer.
func writeOSDetails(w io.Writer) {
	// TODO: Include more details
	var cmdArgs []string
	switch runtime.GOOS {
	case "darwin", "linux":
		cmdArgs = []string{"uname", "-sr"}
	default:
		return
	}
	out, err := exec.Command(cmdArgs[0], cmdArgs[1:]...).CombinedOutput()
	if err == nil {
		fmt.Fprintf(w, "$ uname -sr\n")
		fmt.Fprint(w, string(out)+"\n\n")
	}
}
