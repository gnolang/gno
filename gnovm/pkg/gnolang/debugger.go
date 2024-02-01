package gnolang

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// DebugState is the state of the machine debugger.
type DebugState int

const (
	DebugAtInit DebugState = iota
	DebugAtCmd
	DebugAtRun
	DebugAtExit
)

// Debugger describes a machine debugger state.
type Debugger struct {
	DebugEnabled bool
	DebugAddr    string
	DebugState
	DebugIn  io.ReadCloser
	DebugOut io.Writer

	lastDebugCmd    string
	lastDebugArg    string
	DebugLoc        Location
	PrevDebugLoc    Location
	breakpoints     []Location
	debugCall       []Location // should be provided by machine frame
	debugFrameLevel int
}

type debugCommand struct {
	debugFunc          func(*Machine, string) error // debug command
	usage, short, long string                       // command help texts
}

var debugCmds map[string]debugCommand
var debugCmdNames []string

func init() {
	log.SetFlags(log.Lshortfile)

	// Register debugger commands.
	debugCmds = map[string]debugCommand{
		"break":       {debugBreak, breakUsage, breakShort, breakLong},
		"breakpoints": {debugBreakpoints, breakpointsUsage, breakpointsShort, ""},
		"clear":       {debugClear, clearUsage, clearShort, ""},
		"continue":    {debugContinue, continueUsage, continueShort, ""},
		"detach":      {debugDetach, detachUsage, detachShort, ""},
		"down":        {debugDown, downUsage, downShort, ""},
		"exit":        {debugExit, exitUsage, exitShort, ""},
		"help":        {debugHelp, helpUsage, helpShort, ""},
		"list":        {debugList, listUsage, listShort, listLong},
		"print":       {debugPrint, printUsage, printShort, ""},
		"stack":       {debugStack, stackUsage, stackShort, ""},
		"step":        {debugContinue, stepUsage, stepShort, ""},
		"stepi":       {debugContinue, stepiUsage, stepiShort, ""},
		"up":          {debugUp, upUsage, upShort, ""},
	}

	// Sort command names for help.
	debugCmdNames = make([]string, 0, len(debugCmds))
	for name := range debugCmds {
		debugCmdNames = append(debugCmdNames, name)
	}
	sort.SliceStable(debugCmdNames, func(i, j int) bool { return debugCmdNames[i] < debugCmdNames[j] })

	// Set command aliases.
	debugCmds["b"] = debugCmds["break"]
	debugCmds["bp"] = debugCmds["breakpoints"]
	debugCmds["bt"] = debugCmds["stack"]
	debugCmds["c"] = debugCmds["continue"]
	debugCmds["h"] = debugCmds["help"]
	debugCmds["l"] = debugCmds["list"]
	debugCmds["p"] = debugCmds["print"]
	debugCmds["quit"] = debugCmds["exit"]
	debugCmds["q"] = debugCmds["exit"]
	debugCmds["s"] = debugCmds["step"]
	debugCmds["si"] = debugCmds["stepi"]
}

// Debug is the debug callback invoked at each VM execution step.
func (m *Machine) Debug() {
loop:
	for {
		switch m.DebugState {
		case DebugAtInit:
			initDebugIO(m)
			debugUpdateLoc(m)
			fmt.Fprintln(m.DebugOut, "Welcome to the Gnovm debugger. Type 'help' for list of commands.")
			m.DebugState = DebugAtCmd
		case DebugAtCmd:
			if err := debugCmd(m); err != nil {
				fmt.Fprintln(m.DebugOut, "Command failed:", err)
			}
		case DebugAtRun:
			switch m.lastDebugCmd {
			case "si", "stepi":
				m.DebugState = DebugAtCmd
				debugLineInfo(m)
			case "s", "step":
				if m.DebugLoc != m.PrevDebugLoc {
					m.DebugState = DebugAtCmd
					m.PrevDebugLoc = m.DebugLoc
					debugList(m, "")
					continue loop
				}
			default:
				for _, b := range m.breakpoints {
					if b == m.DebugLoc && m.DebugLoc != m.PrevDebugLoc {
						m.DebugState = DebugAtCmd
						m.PrevDebugLoc = m.DebugLoc
						debugList(m, "")
						continue loop
					}
				}
			}
			break loop
		case DebugAtExit:
			os.Exit(0)
		}
	}
	debugUpdateLoc(m)

	// Keep track of exact locations when performing calls.
	op := m.Ops[m.NumOps-1]
	switch op {
	case OpCall:
		m.debugCall = append(m.debugCall, m.DebugLoc)
	case OpReturn, OpReturnFromBlock:
		m.debugCall = m.debugCall[:len(m.debugCall)-1]
	}
}

// debugCmd processes a debugger REPL command. It displays a prompt, then
// reads and parses a command from the debugger input stream, then executes
// the corresponding function or returns an error.
// If the command is empty, the last non-empty command is repeated.
func debugCmd(m *Machine) error {
	var cmd, arg string
	fmt.Fprint(m.DebugOut, "dbg> ")
	if n, err := fmt.Fscanln(m.DebugIn, &cmd, &arg); errors.Is(err, io.EOF) {
		return debugDetach(m, arg) // Clean close of debugger, the target program resumes.
	} else if n == 0 {
		if m.lastDebugCmd == "" {
			return nil
		}
		cmd, arg = m.lastDebugCmd, m.lastDebugArg
	}
	c, ok := debugCmds[cmd]
	if !ok {
		return errors.New("command not available: " + cmd)
	}
	m.lastDebugCmd, m.lastDebugArg = cmd, arg
	return c.debugFunc(m, arg)
}

// initDebugIO initializes the debugger standard input and output streams.
// If no debug address was specified at program start, the debugger will inherit its
// standard input and output from the process, which will be shared with the target program.
// If the debug address was specified, the program will be blocked until
// a client connection is established. The debugger will use this connection and
// not affecting the target program's.
// An error during connection setting will result in program panic.
func initDebugIO(m *Machine) {
	if m.DebugAddr == "" {
		m.DebugIn = os.Stdin
		m.DebugOut = os.Stdout
		return
	}
	l, err := net.Listen("tcp", m.DebugAddr)
	if err != nil {
		panic(err)
	}
	print("Waiting for debugger client to connect at ", m.DebugAddr)
	conn, err := l.Accept()
	if err != nil {
		panic(err)
	}
	println(" connected!")
	m.DebugIn = conn
	m.DebugOut = conn
}

// debugUpdateLoc computes the source code location for the current VM state.
// The result is stored in Debugger.DebugLoc.
func debugUpdateLoc(m *Machine) {
	loc := m.LastBlock().Source.GetLocation()

	// File must have an unambiguous absolute path.
	if loc.File != "" && !filepath.IsAbs(loc.File) {
		loc.File, _ = filepath.Abs(loc.File)
	}

	if m.DebugLoc.PkgPath == "" ||
		loc.PkgPath != "" && loc.PkgPath != m.DebugLoc.PkgPath ||
		loc.File != "" && loc.File != m.DebugLoc.File {
		m.DebugLoc = loc
	}

	// The location computed from above points to the block start. Examine
	// expressions and statements to have the exact line within the block.

	nx := len(m.Exprs)
	for i := nx - 1; i >= 0; i-- {
		expr := m.Exprs[i]
		if l := expr.GetLine(); l > 0 {
			m.DebugLoc.Line = l
			return
		}
	}

	if len(m.Stmts) > 0 {
		if stmt := m.PeekStmt1(); stmt != nil {
			if l := stmt.GetLine(); l > 0 {
				m.DebugLoc.Line = l
				return
			}
		}
	}
}

// ---------------------------------------
const breakUsage = `break|b [locspec]`
const breakShort = `Set a breakpoint.`
const breakLong = `
The syntax accepted for locspec is:
- <filename>:<line> specifies the line in filename. Filename can be relative.
- <line> specifies the line in the current source file.
- +<offset> specifies the line offset lines after the current one.
- -<offset> specifies the line offset lines before the current one.
`

func debugBreak(m *Machine, arg string) error {
	loc, err := arg2loc(m, arg)
	if err != nil {
		return err
	}
	m.breakpoints = append(m.breakpoints, loc)
	fmt.Fprintf(m.DebugOut, "Breakpoint %d at %s %s:%d\n", len(m.breakpoints)-1, loc.PkgPath, loc.File, loc.Line)
	return nil
}

func arg2loc(m *Machine, arg string) (loc Location, err error) {
	var line int
	loc = m.DebugLoc

	if strings.Contains(arg, ":") {
		// Location is specified by filename:line.
		strs := strings.Split(arg, ":")
		if strs[0] != "" {
			if loc.File, err = filepath.Abs(strs[0]); err != nil {
				return loc, err
			}
			loc.File = filepath.Clean(loc.File)
		}
		if line, err = strconv.Atoi(strs[1]); err != nil {
			return loc, err
		}
		loc.Line = line
		return loc, nil
	}
	if strings.HasPrefix(arg, "+") || strings.HasPrefix(arg, "-") {
		// Location is specified as a line offset from the current line.
		if line, err = strconv.Atoi(arg); err != nil {
			return loc, err
		}
		loc.Line += line
		return loc, nil
	}
	if line, err = strconv.Atoi(arg); err == nil {
		// Location is the line number in the current file.
		loc.Line = line
		return loc, nil
	}
	return loc, err
}

// ---------------------------------------
const breakpointsUsage = `breakpoints|bp`
const breakpointsShort = `Print out info for active breakpoints.`

func debugBreakpoints(m *Machine, arg string) error {
	for i, b := range m.breakpoints {
		fmt.Fprintf(m.DebugOut, "Breakpoint %d at %s %s:%d\n", i, b.PkgPath, b.File, b.Line)
	}
	return nil
}

// ---------------------------------------
const clearUsage = `clear [id]`
const clearShort = `Delete breakpoint (all if no id).`

func debugClear(m *Machine, arg string) error {
	if arg != "" {
		id, err := strconv.Atoi(arg)
		if err != nil || id < 0 || id >= len(m.breakpoints) {
			return fmt.Errorf("invalid breakpoint id: %v", arg)
		}
		m.breakpoints = append(m.breakpoints[:id], m.breakpoints[id+1:]...)
		return nil
	}
	m.breakpoints = nil
	return nil
}

// ---------------------------------------
const continueUsage = `continue|c`
const continueShort = `Run until breakpoint or program termination.`

const stepUsage = `step|s`
const stepShort = `Single step through program.`

const stepiUsage = `stepi|si`
const stepiShort = `Single step a single VM instruction.`

func debugContinue(m *Machine, arg string) error {
	m.DebugState = DebugAtRun
	m.debugFrameLevel = 0
	return nil
}

// ---------------------------------------
const detachUsage = `detach`
const detachShort = `Close debugger and resume program.`

func debugDetach(m *Machine, arg string) error {
	m.DebugEnabled = false
	m.DebugState = DebugAtRun
	m.DebugIn.Close()
	return nil
}

// ---------------------------------------
const downUsage = `down [n]`
const downShort = `Move the current frame down by n (default 1).`

func debugDown(m *Machine, arg string) (err error) {
	n := 1
	if arg != "" {
		if n, err = strconv.Atoi(arg); err != nil {
			return err
		}
	}
	if level := m.debugFrameLevel - n; level >= 0 && level < len(m.debugCall) {
		m.debugFrameLevel = level
	}
	debugList(m, "")
	return nil
}

// ---------------------------------------
const exitUsage = `exit|quit|q`
const exitShort = `Exit the debugger and program.`

func debugExit(m *Machine, arg string) error { m.DebugState = DebugAtExit; return nil }

// ---------------------------------------
const helpUsage = `help|h [command]`
const helpShort = `Print the help message.`

func debugHelp(m *Machine, arg string) error {
	c, ok := debugCmds[arg]
	if !ok && arg != "" {
		return errors.New("command not available")
	}
	if ok {
		t := fmt.Sprintf("%-25s %s", c.usage, c.short)
		if c.long != "" {
			t += "\n" + c.long
		}
		fmt.Fprintln(m.DebugOut, t)
		return nil
	}
	t := "The followings commands are available:\n\n"
	for _, name := range debugCmdNames {
		c := debugCmds[name]
		t += fmt.Sprintf("%-25s %s\n", c.usage, c.short)
	}
	t += "\nType help followed by a command for full documentation."
	fmt.Fprintln(m.DebugOut, t)
	return nil
}

// ---------------------------------------
const listUsage = `list|l [locspec]`
const listShort = `Show source code.`
const listLong = `
See 'help break' for locspec syntax. If locspec is empty,
list shows the source code around the current line.
`

func debugList(m *Machine, arg string) (err error) {
	loc := m.DebugLoc
	hideCursor := false

	if arg == "" {
		debugLineInfo(m)
		if m.lastDebugCmd == "up" || m.lastDebugCmd == "down" {
			loc = debugFrameLoc(m, m.debugFrameLevel)
			fmt.Fprintf(m.DebugOut, "Frame %d: %s:%d\n", m.debugFrameLevel, loc.File, loc.Line)
		}
	} else {
		if loc, err = arg2loc(m, arg); err != nil {
			return err
		}
		hideCursor = true
		fmt.Fprintf(m.DebugOut, "Showing %s:%d\n", loc.File, loc.Line)
	}
	file, line := loc.File, loc.Line
	lines, offset, err := sourceLines(file, line)
	if err != nil {
		return err
	}
	for i, l := range lines {
		cursor := ""
		if !hideCursor && file == loc.File && loc.Line == i+offset {
			cursor = "=>"
		}
		fmt.Fprintf(m.DebugOut, "%2s %4d: %s\n", cursor, i+offset, l)
	}
	return nil
}

func debugLineInfo(m *Machine) {
	line := string(m.Package.PkgName)
	if len(m.Frames) > 0 {
		f := m.Frames[len(m.Frames)-1]
		if f.Func != nil {
			line += "." + string(f.Func.Name) + "()"
		}
	}
	fmt.Fprintf(m.DebugOut, "> %s %s:%d\n", line, m.DebugLoc.File, m.DebugLoc.Line)
}

const listLength = 10 // number of lines to display

func sourceLines(name string, n int) ([]string, int, error) {
	buf, err := os.ReadFile(name)
	if err != nil {
		return nil, 1, err
	}
	lines := strings.Split(string(buf), "\n")
	start := max(1, n-listLength/2) - 1
	end := min(start+listLength, len(lines))
	return lines[start:end], start + 1, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ---------------------------------------
const printUsage = `print|p <expression>`
const printShort = `Print a variable or expression.`

func debugPrint(m *Machine, arg string) error {
	if arg == "" {
		return errors.New("missing argument")
	}
	b := m.Blocks[len(m.Blocks)-1-m.debugFrameLevel]
	for i, name := range b.Source.GetBlockNames() {
		// TODO: handle index and selector expressions.
		if string(name) == arg {
			fmt.Fprintln(m.DebugOut, b.Values[i])
			return nil
		}
	}
	return fmt.Errorf("could not find symbol value for %s", arg)
}

// ---------------------------------------
const stackUsage = `stack|bt`
const stackShort = `Print stack trace.`

func debugStack(m *Machine, arg string) error {
	l := len(m.Frames) - 1
	// List stack frames in reverse array order. Deepest level is 0.
	for i := l; i >= 0; i-- {
		f := m.Frames[i]
		loc := debugFrameLoc(m, l-i)
		fmt.Fprintf(m.DebugOut, "%d\tin %s.%s\n\tat %s:%d\n", l-i, f.LastPackage.PkgPath, f.Func, loc.File, loc.Line)
	}
	return nil
}

func debugFrameLoc(m *Machine, n int) Location {
	if n == 0 || len(m.debugCall) == 0 {
		return m.DebugLoc
	}
	return m.debugCall[len(m.debugCall)-n]
}

// ---------------------------------------
const upUsage = `up [n]`
const upShort = `Move the current frame up by n (default 1).`

func debugUp(m *Machine, arg string) (err error) {
	n := 1
	if arg != "" {
		if n, err = strconv.Atoi(arg); err != nil {
			return err
		}
	}
	if level := m.debugFrameLevel + n; level >= 0 && level < len(m.debugCall) {
		m.debugFrameLevel = level
	}
	debugList(m, "")
	return nil
}
