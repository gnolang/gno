package profiler

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// ProfilerCLI manages the interactive profiler interface
type ProfilerCLI struct {
	in      io.Reader      // input stream, defaults to Stdin
	out     io.Writer      // output stream, defaults to Stdout
	scanner *bufio.Scanner // to parse input per line
	profile *Profile       // current profile data
	store   Store          // for accessing source files
	lastCmd string         // last command executed
	lastArg string         // last command arguments

	// Filtering and focusing
	focusFunc   string   // function to focus on
	hideFunc    []string // functions to hide
	ignoreFunc  []string // functions to ignore
	minSamples  int64    // minimum samples to show
	sampleIndex int      // which sample value to use (for multiple sample types)

	// Display options
	cumulative bool   // show cumulative values
	flat       bool   // show flat values only
	addresses  bool   // show addresses
	lines      bool   // show line numbers
	nodeCount  int    // max number of nodes to show
	unit       string // display unit (cycles, ms, etc)
}

// ProfilerCommand describes a profiler command
type profilerCommand struct {
	execFunc           func(*ProfilerCLI, string) error // command function
	usage, short, long string                           // help texts
}

var (
	profilerCmds     map[string]profilerCommand
	profilerCmdNames []string
)

func init() {
	// Register profiler commands
	profilerCmds = map[string]profilerCommand{
		"top":       {profilerTop, "top [n]", "Show top N functions", topLong},
		"list":      {profilerList, "list <function>", "Show annotated source for function", pprofListLong},
		"tree":      {profilerTree, "tree [function]", "Show call tree", treeLong},
		"focus":     {profilerFocus, "focus <function>", "Focus on specific function", focusLong},
		"ignore":    {profilerIgnore, "ignore <function>", "Ignore function in output", ""},
		"hide":      {profilerHide, "hide <function>", "Hide function from output", ""},
		"show":      {profilerShow, "show", "Show current focus/ignore/hide settings", ""},
		"reset":     {profilerReset, "reset", "Reset all focus/ignore/hide settings", ""},
		"sample":    {profilerSample, "sample <index>", "Select sample type", sampleLong},
		"save":      {profilerSave, "save <file>", "Save profile to file", ""},
		"help":      {profilerHelp, "help [command]", "Show help", ""},
		"quit":      {profilerQuit, "quit", "Exit profiler", ""},
		"cum":       {profilerToggleCum, "cum", "Toggle cumulative mode", ""},
		"flat":      {profilerToggleFlat, "flat", "Toggle flat mode", ""},
		"addresses": {profilerToggleAddr, "addresses", "Toggle showing addresses", ""},
		"lines":     {profilerToggleLines, "lines", "Toggle showing line numbers", ""},
		"nodecount": {profilerNodeCount, "nodecount <n>", "Set max nodes to show", ""},
		"unit":      {profilerUnit, "unit <unit>", "Set display unit", unitLong},
	}

	// Sort command names for help
	profilerCmdNames = make([]string, 0, len(profilerCmds))
	for name := range profilerCmds {
		profilerCmdNames = append(profilerCmdNames, name)
	}
	sort.Strings(profilerCmdNames)

	// Set command aliases
	profilerCmds["t"] = profilerCmds["top"]
	profilerCmds["l"] = profilerCmds["list"]
	profilerCmds["h"] = profilerCmds["help"]
	profilerCmds["q"] = profilerCmds["quit"]
	profilerCmds["exit"] = profilerCmds["quit"]
}

// Long help texts
const (
	topLong = `Show the top N functions by sample count.
Default is 10 if N is not specified.

Examples:
  top        Show top 10 functions
  top 20     Show top 20 functions
  top -cum   Show top 10 by cumulative count`

	pprofListLong = `Show annotated source code for a function.
The function name can be a partial match.

Examples:
  list main.main     Show source for main.main
  list String        Show source for functions containing "String"
  list .             Show source for current function`

	treeLong = `Show the call tree for the profile.
If a function is specified, show the call tree rooted at that function.

Examples:
  tree              Show full call tree
  tree main.main    Show call tree starting from main.main`

	focusLong = `Focus on samples containing the specified function.
Only samples that include this function in their stack will be shown.

Examples:
  focus main.main   Focus on samples containing main.main
  focus String      Focus on samples with functions containing "String"`

	sampleLong = `Select which sample type to display.
Most profiles have multiple sample types (e.g., count and cycles).

Examples:
  sample 0    Use first sample type (usually count)
  sample 1    Use second sample type (usually cycles)`

	unitLong = `Set the unit for displaying sample values.

Examples:
  unit cycles    Show values in cycles
  unit ms        Show values in milliseconds
  unit auto      Automatically choose unit`
)

// NewProfilerCLI creates a new interactive profiler CLI
func NewProfilerCLI(profile *Profile, store Store) *ProfilerCLI {
	return &ProfilerCLI{
		in:         os.Stdin,
		out:        os.Stdout,
		profile:    profile,
		store:      store,
		nodeCount:  10,
		cumulative: true,
		unit:       "auto",
		hideFunc:   make([]string, 0),
		ignoreFunc: make([]string, 0),
	}
}

// Run starts the interactive profiler session
func (p *ProfilerCLI) Run() error {
	p.scanner = bufio.NewScanner(p.in)

	fmt.Fprintln(p.out, "Entering interactive pprof mode.")
	fmt.Fprintf(p.out, "Type 'help' for commands, 'quit' to exit.\n")

	// Show initial top functions
	if err := p.execCommand("top", ""); err != nil {
		fmt.Fprintf(p.out, "Error: %v\n", err)
	}

	for {
		fmt.Fprint(p.out, "(pprof) ")

		if !p.scanner.Scan() {
			return nil
		}

		line := strings.TrimSpace(p.scanner.Text())
		if line == "" {
			// Repeat last command
			line = p.lastCmd
			if p.lastArg != "" {
				line += " " + p.lastArg
			}
		}

		parts := strings.SplitN(line, " ", 2)
		cmd := parts[0]
		arg := ""
		if len(parts) > 1 {
			arg = strings.TrimSpace(parts[1])
		}

		if err := p.execCommand(cmd, arg); err != nil {
			if err == io.EOF {
				return nil
			}
			fmt.Fprintf(p.out, "Error: %v\n", err)
		}
	}
}

// execCommand executes a profiler command
func (p *ProfilerCLI) execCommand(cmd, arg string) error {
	c, ok := profilerCmds[cmd]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd)
	}

	p.lastCmd = cmd
	p.lastArg = arg

	return c.execFunc(p, arg)
}

// SetInput sets the input reader
func (p *ProfilerCLI) SetInput(r io.Reader) {
	p.in = r
}

// SetOutput sets the output writer
func (p *ProfilerCLI) SetOutput(w io.Writer) {
	p.out = w
}
