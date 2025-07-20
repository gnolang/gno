// Package main implements a CLI tool to parse Makefile targets and scan directories.
// It extracts targets (lines beginning with a letter and containing a colon),
// captures inline comments (starting with one or more '#'), and optionally
// expands '%' targets with wildcard values. It also scans provided directories
// for Makefiles, flags those with a "help" target, and reads the first line of
// any README.md there as a banner.
//
// The --invocation-dir-prefix option is used solely to improve help output clarity
// when Make is invoked with -C, since subshell execution resets PWD and otherwise
// loses this context.
//
// Usage:
//
//	go run main.go [OPTIONS] <Makefile>
//
// Options:
//
//	--invocation-dir-prefix PATH  Path prefix to reflect use of `make -C DIR`. Used
//			                      only to adjust help output paths so they match
//			                      how Make was originally invoked by the user.
//	--dirs DIR [...]              List of directories to scan for Makefiles
//	--wildcard WILD [...]         List of wildcard substitutions for '%' targets
//	-h, --help                    Show usage and exit
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

var (
	numPrefixRE        = regexp.MustCompile(`^(\d+):`)
	percentNumPrefixRE = regexp.MustCompile(`^%(\d+):`)
)

// flagList accumulates multiple flag values into a slice.
type flagList []string

// String returns the comma‑separated flag values.
func (f *flagList) String() string {
	strArrVal := []string(*f)
	byteArr, err := json.Marshal(strArrVal)
	if err != nil {
		// This shouldn't be possible
		panic(err)
	}
	return string(byteArr)
}

// Set appends a new value to the flag list.
func (f *flagList) Set(val string) error {
	*f = append(*f, val)
	return nil
}

// flagList accumulates multiple flag values into a slice.
type nestedFlagList [][]string

// String returns the comma‑separated flag values.
func (nf *nestedFlagList) String() string {
	strArrArrVal := [][]string(*nf)
	byteArr, err := json.Marshal(strArrArrVal)
	if err != nil {
		// This shouldn't be possible
		panic(err)
	}
	return string(byteArr)
}

// Set appends a new value to the flag list.
func (nf *nestedFlagList) Set(val string) error {
	targetIdx := 0
	rest := val

	if m := numPrefixRE.FindStringSubmatch(val); m != nil {
		idx, err := strconv.Atoi(m[1])
		if err != nil || idx <= 0 {
			// This should not be reachable.
			return fmt.Errorf("invalid prefix index in %q", val)
		}
		targetIdx = idx - 1
		rest = val[len(m[0]):]
	}

	// Ensure enough inner slices
	for len(*nf) <= targetIdx {
		*nf = append(*nf, []string{})
	}

	(*nf)[targetIdx] = append((*nf)[targetIdx], rest)
	return nil
}

// Config holds all command‑line options.
type Config struct {
	Makefile   string         // path to the main Makefile
	RelativeTo string         // base path for printing sub‑directory commands
	Dirs       flagList       // additional directories to scan
	Wildcards  nestedFlagList // wildcard values to expand '%' targets
}

// parseConfig parses flags and the single required Makefile argument.
// Returns a non‑nil *Config on success, or an error on bad input.
func parseConfig(args []string) (*Config, *flag.FlagSet, error) {
	cfg := &Config{}
	fs := flag.NewFlagSet("makefile-help", flag.ContinueOnError)

	fs.StringVar(&cfg.RelativeTo, "invocation-dir-prefix", "",
		"path prefix to reflect use of `make -C DIR`. Used only to adjust help output paths to match how Make was originally invoked by the user.")
	fs.Var(&cfg.Dirs, "dir", "directory to scan for Makefiles (repeatable)")
	fs.Var(&cfg.Wildcards, "wildcard", "value to substitute for '%' in targets (repeatable)")

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
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	targets := make(map[string]string)
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimRightFunc(scanner.Text(), unicode.IsSpace)

		// Skip empty lines, non-letter starts, lines without ':', legacy lines, or variable assignments
		if len(line) == 0 || !unicode.IsLetter(rune(line[0])) {
			continue
		}

		colon := strings.IndexByte(line, ':')
		if colon == -1 || (colon+1 < len(line) && line[colon+1] == '=') {
			continue
		}

		name := line[:colon]
		desc := ""

		// Find description after first single '#'.
		// Skip lines marked @LEGACY.
		if hashPos := strings.IndexByte(line, '#'); hashPos > colon {
			for (hashPos+1 < len(line)) && (line[hashPos+1] == '#') {
				hashPos += 1
			}
			desc = strings.TrimSpace(line[hashPos+1:])
			if strings.Contains(desc, "@LEGACY") {
				continue
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
func maxKeyLength(keys []string, wildcards [][]string) int {
	totalWildcards := 0
	for _, w := range wildcards {
		totalWildcards += len(w)
	}
	flatWild := make([]string, 0, totalWildcards)
	for _, wo := range wildcards {
		for _, wi := range wo {
			flatWild = append(flatWild, wi)
		}
	}
	wildMax := maxStringLength(flatWild)
	maxLen := 0
	for _, k := range keys {
		length := len(k)
		if strings.Contains(k, "%") && len(flatWild) > 0 {
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
func printTargets(w io.Writer, targets map[string]string, wildcards [][]string, banners map[string]string) {
	keys := make([]string, 0, len(targets))
	for k := range targets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i := range wildcards {
		sort.Strings(wildcards[i])
	}

	width := maxKeyLength(keys, wildcards)
	for _, key := range keys {
		desc := targets[key]
		wildIdx := 0

		if m := percentNumPrefixRE.FindStringSubmatch(desc); m != nil {
			idx, err := strconv.Atoi(m[1])
			if err != nil || idx <= 0 {
				// This should not be reachable.
				panic(fmt.Sprintf("invalid prefix index in %q", desc))
			}
			wildIdx = idx - 1
			desc = desc[len(m[0]):]
		}

		if strings.Contains(key, "%") && (len(wildcards) > wildIdx) && (len(wildcards[wildIdx]) > 0) {
			for _, wild := range wildcards[wildIdx] {
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

func run(args []string, stdout, stderr io.Writer) error {
	cfg, fs, err := parseConfig(args)
	flag.CommandLine = fs
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	if len(cfg.Wildcards) < 1 {
		cfg.Wildcards = [][]string{{}}
	}

	banners := scrapeReadmeBanners(cfg.Wildcards[0], cfg.Dirs)
	fmt.Fprintln(stdout, "Available make targets:")
	tMap, err := extractMakefileTargets(cfg.Makefile)
	if err != nil {
		return fmt.Errorf("failed to parse Makefile: %w", err)
	}
	printTargets(stdout, tMap, cfg.Wildcards, banners)
	printSubdirs(stdout, cfg.RelativeTo, cfg.Dirs, banners)
	return nil
}

func ErrorToExitCode(err error, stderr io.Writer) int {
	if err != nil {
		fmt.Fprintln(stderr, "Error:", err.Error())
		flag.PrintDefaults()
		return 1
	}
	return 0
}

func main() {
	os.Exit(ErrorToExitCode(run(os.Args[1:], os.Stdout, os.Stderr), os.Stderr))
}
