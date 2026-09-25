package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gh-report:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		dataDir = flag.String("data", "data", "directory holding fetched JSON files")
		ansiF   = flag.Bool("ansi", false, "render with ANSI colors (terminal)")
		jsonF   = flag.Bool("json", false, "render as JSON")
		repoF   = flag.String("repo", "", "restrict to one owner/repo")
	)
	flag.Parse()

	files, err := filepath.Glob(filepath.Join(*dataDir, "*--*.json"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "no data, run `make fetch` first")
		return nil
	}

	var allEntries []Entry
	for _, f := range files {
		repo := filenameToRepo(f)
		if *repoF != "" && repo != *repoF {
			continue
		}
		body, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", f, err)
			continue
		}
		entries, err := LoadRepoJSON(repo, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", f, err)
			continue
		}
		allEntries = append(allEntries, entries...)
	}

	// Sort: stable order by UpdatedAt descending within each future section.
	sort.SliceStable(allEntries, func(i, j int) bool {
		return allEntries[i].UpdatedAt.After(allEntries[j].UpdatedAt)
	})

	r := Classify(allEntries)
	if len(r.Sections) == 0 {
		fmt.Println("no items in window")
		return nil
	}
	switch {
	case *jsonF:
		return RenderJSON(os.Stdout, r)
	case *ansiF:
		return RenderANSI(os.Stdout, r)
	default:
		return RenderMarkdown(os.Stdout, r)
	}
}

// filenameToRepo turns "data/gnolang--gno.json" into "gnolang/gno".
func filenameToRepo(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), ".json")
	return strings.Replace(base, "--", "/", 1)
}
