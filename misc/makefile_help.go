// Package main implements a CLI tool to parse Makefile targets and scan directories.
// It extracts targets (lines beginning with a letter and containing a colon),
// captures inline comments (starting with one or more '#'), and optionally
// expands '%' targets with wildcard values. It also scans provided directories
// for Makefiles, flags those with a "help" target, and reads the first line of
// any README.md there as a banner.
//
// Usage:
//
//	go run main.go [OPTIONS] <Makefile>
//
// Options:
//
//	-r, --relative-to PATH    Treat PATH as the relative invocation path
//	-d, --dirs DIR [...]      List of directories to scan for Makefiles
//	-w, --wildcard WILD [...] List of wildcard substitutions for '%' targets
//	-h, --help                Show usage and exit
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// flagList accumulates multiple flag values into a slice.
type flagList []string

// String returns the comma‑separated flag values.
func (f *flagList) String() string {
	return strings.Join(*f, ",")
}

// Set appends a new value to the flag list.
func (f *flagList) Set(val string) error {
	*f = append(*f, val)
	return nil
}

// Config holds all command‑line options.
type Config struct {
	Makefile   string   // path to the main Makefile
	RelativeTo string   // base path for printing sub‑directory commands
	Dirs       flagList // additional directories to scan
	Wildcards  flagList // wildcard values to expand '%' targets
}

// parseConfig parses flags and the single required Makefile argument.
// Returns a non‑nil *Config on success, or an error on bad input.
func parseConfig(args []string) (*Config, *flag.FlagSet, error) {
	cfg := &Config{}
	fs := flag.NewFlagSet("makefile-help", flag.ContinueOnError)

	fs.StringVar(&cfg.RelativeTo, "relative-to", "", "base path for sub‑directory commands")
	fs.StringVar(&cfg.RelativeTo, "r", "", "shorthand for --relative-to")
	fs.Var(&cfg.Dirs, "dir", "directory to scan for Makefiles (repeatable)")
	fs.Var(&cfg.Dirs, "d", "shorthand for --dir")
	fs.Var(&cfg.Wildcards, "wildcard", "value to substitute for '%' in targets (repeatable)")
	fs.Var(&cfg.Wildcards, "w", "shorthand for --wildcard")

	if err := fs.Parse(args); err != nil {
		return nil, fs, err
	}
	rest := fs.Args()
	if len(rest) != 1 {
		return nil, fs, errors.New("must specify exactly one Makefile path")
	}
	cfg.Makefile = rest[0]
	info, err := os.Stat(cfg.Makefile)
	if err != nil || info.IsDir() {
		return nil, fs, fmt.Errorf("cannot read Makefile %q: %w", cfg.Makefile, err)
	}
	return cfg, fs, nil
}

// extractMakefileTargets reads filePath and returns a map of target⇒description.
// It ignores lines not starting with a letter, without ':', or marked @LEGACY.
// Inline comments after one or more '#'s become descriptions.
func extractMakefileTargets(filePath string) (map[string]string, error) {
	legacyRe := regexp.MustCompile(`#.*@LEGACY\b`)

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	targets := make(map[string]string)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || !unicode.IsLetter(rune(line[0])) {
			continue
		}
		colon := strings.IndexRune(line, ':')
		if colon < 0 || legacyRe.MatchString(line) {
			continue
		}

		name := line[:colon]
		desc := ""
		for i := colon + 1; i < len(line); i++ {
			if line[i] == '#' && (i+1 < len(line) && line[i+1] != '#') {
				desc = strings.TrimSpace(line[i+1:])
				break
			}
		}
		targets[name] = desc
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return targets, nil
}

// maxStringLength returns the length of the longest string in items.
func maxStringLength(items []string) int {
	maxLen := 0
	for _, s := range items {
		if len(s) > maxLen {
			maxLen = len(s)
		}
	}
	return maxLen
}

// maxKeyLength calculates the column width for printing keys,
// accounting for '%' expansions if wildcards are provided.
func maxKeyLength(keys, wildcards []string) int {
	wildMax := maxStringLength(wildcards)
	maxLen := 0
	for _, k := range keys {
		length := len(k)
		if strings.Contains(k, "%") && len(wildcards) > 0 {
			length += wildMax - 1
		}
		if length > maxLen {
			maxLen = length
		}
	}
	return maxLen
}

// readReadmeBanner finds and returns a parenthesized summary
// from the first non‑empty line of README.md in dir, or "" if none.
func readReadmeBanner(dir string) (string, error) {
	// strip leading "#", spaces, and an optional "dir:" prefix
	prefixRe := regexp.MustCompile(`(?i)^ *(` +
		regexp.QuoteMeta(dir) +
		`|` + "`" + regexp.QuoteMeta(dir) + "`" +
		`) *((-+|:) *|$)`)

	path := filepath.Join(dir, "README.md")
	data, err := os.ReadFile(path)
	if err != nil {
		// missing or unreadable → no banner
		return "", nil
	}
	line := strings.SplitN(string(data), "\n", 2)[0]
	line = strings.TrimSpace(strings.TrimLeft(line, "# "))
	line = prefixRe.ReplaceAllString(line, "")
	if line == "" {
		return "", nil
	}
	return fmt.Sprintf(" (%s)", line), nil
}

// scrapeReadmeBanners returns a map[dir]banner for each dir in wildcards or dirs.
func scrapeReadmeBanners(wildcards, dirs []string) map[string]string {
	banners := make(map[string]string)
	for _, group := range [][]string{wildcards, dirs} {
		for _, d := range group {
			if _, seen := banners[d]; !seen {
				b, _ := readReadmeBanner(d)
				banners[d] = b
			}
		}
	}
	return banners
}

// printTargets lists all targets and descriptions, expanding '%' with wildcards.
func printTargets(w io.Writer, targets map[string]string, wildcards []string, banners map[string]string) {
	keys := make([]string, 0, len(targets))
	for k := range targets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sort.Strings(wildcards)

	width := maxKeyLength(keys, wildcards)
	for _, key := range keys {
		desc := targets[key]
		if strings.Contains(key, "%") && len(wildcards) > 0 {
			for _, wild := range wildcards {
				ek := strings.ReplaceAll(key, "%", wild)
				if desc == "" {
					fmt.Fprintf(w, "  %s\n", ek)
				} else {
					ed := strings.ReplaceAll(desc, "%", wild)
					fmt.Fprintf(w, "  %-*s   <-- %s%s\n", width, ek, ed, banners[wild])
				}
			}
		} else {
			if desc == "" {
				fmt.Fprintf(w, "  %s\n", key)
			} else {
				fmt.Fprintf(w, "  %-*s   <-- %s\n", width, key, desc)
			}
		}
	}
}

// printSubdirs scans each dir for a Makefile, marks '*' if it has a 'help' target,
// and prints a make -C invocation with any README banner.
func printSubdirs(w io.Writer, relativeTo string, dirs []string, banners map[string]string) {
	if len(dirs) == 0 {
		return
	}
	fmt.Fprintln(w, "\nSub‑directories with make targets:")
	sort.Strings(dirs)

	noteHelp := false
	for _, d := range dirs {
		mf := filepath.Join(d, "Makefile")
		info, err := os.Stat(mf)
		if err != nil || info.IsDir() {
			continue
		}
		tMap, err := extractMakefileTargets(mf)
		hasHelp := false
		if err == nil {
			_, ok := tMap["help"]
			hasHelp = ok
		}
		star := " "
		if hasHelp {
			star = "*"
			noteHelp = true
		}
		targetDir := d
		if relativeTo != "" {
			targetDir = filepath.ToSlash(filepath.Join(relativeTo, d))
		}
		fmt.Fprintf(w, "    %s  make -C %s%s\n", star, targetDir, banners[d])
	}
	if noteHelp {
		fmt.Fprintln(w, "\n       * Is documented with a `help` target.")
	}
}

func run(args []string, stdout, stderr io.Writer) int {
	cfg, fs, err := parseConfig(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintln(stderr, "Error:", err)
		fs.PrintDefaults()
		return 1
	}

	banners := scrapeReadmeBanners(cfg.Wildcards, cfg.Dirs)
	fmt.Fprintln(stdout, "Available make targets:")
	tMap, err := extractMakefileTargets(cfg.Makefile)
	if err != nil {
		fmt.Fprintln(stderr, "Failed to parse Makefile:", err)
		return 2
	}
	printTargets(stdout, tMap, cfg.Wildcards, banners)
	printSubdirs(stdout, cfg.RelativeTo, cfg.Dirs, banners)
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
