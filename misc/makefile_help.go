// Package main implements a CLI tool to parse Makefile targets and scan directories.
// It extracts targets (lines beginning with a letter and containing a colon),
// captures inline comments (only '#' not followed by whitespace), and optionally
// expands '%' targets with wildcard values. It also scans provided directories
// for Makefiles, flags those with a "help" target, and reads the first line of
// any README.md there as a banner.
//
// Usage:
//   go run main.go [OPTIONS] <Makefile>
//
// Options:
//   -r, --relative-to PATH    Treat PATH as the relative invocation path
//   -d, --dirs DIR [...]      List of directories to scan for Makefiles
//   -w, --wildcard WILD [...] List of wildcard substitutions for '%' targets
//   -h, --help                Show usage and exit

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

type stringList struct {
    vals    []string
}

func (s *stringList) String() string {
    return strings.Join(s.vals, ",")
}

func (s *stringList) Set(val string) error {
    s.vals = append(s.vals, val)
    return nil
}

// Config holds command-line options.
type Config struct {
	Makefile    string
	RelativeTo  string
	Dirs        stringList
	Wildcards   stringList
}

// parseArgs parses CLI arguments from os.Args[1:].
func parseArgs(args []string) (Config, *flag.FlagSet, error) {
	cfg := Config{}
	// showHelp := false
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	// fs.BoolVar(&showHelp, "help", false, "show help")
	fs.StringVar(&cfg.RelativeTo, "relative-to", "", "relative-to path")
	fs.Var(&cfg.Dirs, "dir", "directory to scan for Makefiles (repeatable)")
	fs.Var(&cfg.Wildcards, "wildcard", "wildcard substitution (repeatable)")

	if err := fs.Parse(args); err != nil {
		return cfg, fs, err
	}
	rest := fs.Args()
	if len(rest) != 1 {
		return cfg, fs, errors.New("Expected exactly one makefile")  
  }
  cfg.Makefile = rest[0]
	if _, err := os.Stat(cfg.Makefile); err != nil {
		return cfg, fs, fmt.Errorf("cannot read Makefile '%s': %w", cfg.Makefile, err)
	}
	return cfg, fs, nil
}

// extractTargets reads the Makefile and returns a map[target]description.
func extractTargets(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	targets := make(map[string]string)
	s := bufio.NewScanner(file)
	for s.Scan() {
		line := s.Text()
		// must start with letter and contain ':'
		if len(line) == 0 || !unicode.IsLetter(rune(line[0])) {
			continue
		}
		ci := strings.IndexRune(line, ':')
		if ci < 0 {
			continue
		}
		// exclude LEGACY
		detectLegacyPat := regexp.MustCompile("#.*@LEGACY\\b")
		if detectLegacyPat.MatchString(line) {
			continue
		}
		name := line[:ci]
		// find '#' not followed by another '#'
		descIdx := -1
		for i := ci + 1; i < len(line); i++ {
			if line[i] == '#' && (i+1 < len(line) && line[i+1] != '#') {
				descIdx = i
				break
			}
		}
		desc := ""
		if descIdx >= 0 {
			d := line[descIdx+1:]
			desc = strings.TrimSpace(d)
		}
		targets[name] = desc
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return targets, nil
}

// maxStringLength returns the max length among str.
func maxStringLength(str []string) int {
	max := 0
	for _, s := range str {
		if l := len(s); l > max {
			max = l
		}
	}
	return max
}

// maxKeyLength returns the max length among keys.
func maxKeyLength(keys, wilds []string) int {
	maxWildLen := maxStringLength(wilds)
	max := 0
	for _, k := range keys {
		l := len(k)
		if (len(wilds) > 0) && strings.Contains(k, "%") {
			l += (maxWildLen - 1)
		}
		if l > max {
			max = l
		}
	}
	return max
}

func scrapeReadmeBanners(wild,dirs []string) map[string]string {
	dirBanners := make(map[string]string)
	addBanner := func (list []string) {
		for _,dirName := range list {
			_, found := dirBanners[dirName]
			if !found {
				banner, _ := readReadmeBanner(dirName)
				dirBanners[dirName] = banner
			}
		}
	}
	addBanner(wild)
	addBanner(dirs)
	return dirBanners
}

// readReadmeBanner returns the first line of README.md in dir, parenthesized.
func readReadmeBanner(dir string) (string, error) {
	p := filepath.Join(dir, "README.md")
	b, err := os.ReadFile(p)
	if err != nil {
		return "", nil // treat missing/unreadable as no banner
	}
	line := strings.SplitN(string(b), "\n", 2)[0]
	// strip leading hashes/spaces
	line = strings.TrimLeft(line, " #")
	line = strings.TrimSpace(line)
  removeNamePrefixPat := regexp.MustCompile("(?i)^ *(" + regexp.QuoteMeta(dir) + "|`" + regexp.QuoteMeta(dir) + "`) *((--*|:) *|$)")
  line = removeNamePrefixPat.ReplaceAllString(line,"")
	if line == "" {
		return "", nil
	}
	return fmt.Sprintf(" (%s)", line), nil
}

// displayTargets prints targets and comments, with wildcard expansion.
func displayTargets(targetDescs map[string]string, wilds []string, dirBanners map[string]string) {
	// gather keys
	keys := make([]string, 0, len(targetDescs))
	for k := range targetDescs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sort.Strings(wilds)
	width := maxKeyLength(keys, wilds)
	// non-% targets
	for _, k := range keys {
		if !strings.Contains(k, "%") || (len(wilds) < 1) {
			if targetDescs[k] == "" {
				fmt.Printf("  %s\n", k)
			} else {
				fmt.Printf("  %-*s   <-- %s\n", width, k, targetDescs[k])
			}
		} else {
			for _, w := range wilds {
				kExpanded := strings.ReplaceAll(k,"%",w)
				if targetDescs[k] == "" {
					fmt.Printf("  %s\n", kExpanded)
				} else {
					descExpanded := strings.ReplaceAll(targetDescs[k],"%",w)
					fmt.Printf("  %-*s   <-- %s%s\n", width, kExpanded, descExpanded, dirBanners[w])
				}
			}
		}
	}
}

// displayDirs prints provided directories if they contain Makefile.
func displayDirs(relDir string, dirs []string, dirBanners map[string]string) {
	if len(dirs) < 1 {
		return
	}
	fmt.Println()
	fmt.Println("Sub-directories with make targets:")
	sort.Strings(dirs)
	helpTargetFound := false
	for _, d := range dirs {
		mf := filepath.Join(d, "Makefile")
		if fi, err := os.Stat(mf); err == nil && !fi.IsDir() {
			// check help target
			hasHelpTarget := false
			if tmap, err := extractTargets(mf); err == nil {
				if _, ok := tmap["help"]; ok {
					hasHelpTarget = true
				}
			}
			star := " "
			if hasHelpTarget {
				star = "*"
				helpTargetFound = true
			}
			dispDir := d
			if relDir != "" {
				dispDir = relDir + "/" + d
			}
			banner := dirBanners[d]
			fmt.Printf("    %s  make -C %s%s\n", star, dispDir, banner)
		}
	}
	if helpTargetFound {
		fmt.Printf("\n       * Is documented with a `help` target.\n")
	}
}

func main() {
	cfg, flagSet, err := parseArgs(os.Args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			// flagSet.PrintDefaults() // already done
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, "Error:", err)
		flagSet.PrintDefaults()
		os.Exit(1)
	}
	dirBanners := scrapeReadmeBanners(cfg.Wildcards.vals, cfg.Dirs.vals)
	fmt.Println("Available make targets:")
	targetMap, err := extractTargets(cfg.Makefile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse Makefile:", err)
		os.Exit(2)
	}
	displayTargets(targetMap, cfg.Wildcards.vals, dirBanners)
	displayDirs(cfg.RelativeTo, cfg.Dirs.vals, dirBanners)
}
