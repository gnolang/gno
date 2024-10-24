package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

// newVersionCmd creates a new version command
func newVersionCmd(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "version",
			ShortUsage: "version",
			ShortHelp:  "Display installed gno version",
		},
		nil,
		func(_ context.Context, args []string) error {
			version, err := getGnoVersion()
			if err != nil {
				io.ErrPrintln("error retrieving version:", err)
				return err
			}
			io.Println("gno version:", version)
			return nil
		},
	)
}

// getGnoVersion attempts to retrieve the Gno version using different methods
func getGnoVersion() (string, error) {
	if version, err := getGoPathVersion(); err == nil {
		return version, nil
	}
	if version, err := getGitVersion(); err == nil {
		return version, nil
	}
	return "unknown", fmt.Errorf("unable to determine gno version")
}

// getGoPathVersion looks for the version in GOPATH
func getGoPathVersion() (string, error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", fmt.Errorf("GOPATH not set")
	}

	gnoPath := filepath.Join(gopath, "pkg", "mod", "github.com", "gnolang", "gno@*")
	matches, err := filepath.Glob(gnoPath)
	if err != nil || len(matches) == 0 {
		return "", fmt.Errorf("gno version not found in GOPATH")
	}

	version := filepath.Base(matches[len(matches)-1])
	version = strings.TrimPrefix(version, "gno@")
	return version, nil
}

// getGitVersion retrieves the version from git repository
func getGitVersion() (string, error) {
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		return "", fmt.Errorf("git repository not found")
	}

	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err = cmd.Output()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("git-%s", strings.TrimSpace(string(output))), nil
}
