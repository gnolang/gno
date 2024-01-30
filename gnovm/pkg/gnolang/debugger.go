package gnolang

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
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

	lastDebugCmd string
	lastDebugArg string
	DebugLoc     Location
	PrevDebugLoc Location
	breakpoints  []Location
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
		"exit":        {debugExit, exitUsage, exitShort, ""},
		"help":        {debugHelp, helpUsage, helpShort, ""},
		"list":        {debugList, listUsage, listShort, ""},
		"print":       {debugPrint, printUsage, printShort, ""},
		"step":        {debugContinue, stepUsage, stepShort, ""},
		"stepi":       {debugContinue, stepiUsage, stepiShort, ""},
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
	var filename string
	var line int

	loc = m.DebugLoc
	if strings.Contains(arg, ":") {
		// Location is specified by filename:line or function:line
		strs := strings.Split(arg, ":")
		filename = strs[0]
		if line, err = strconv.Atoi(strs[1]); err != nil {
			return loc, err
		}
		if loc.File, err = filepath.Abs(filename); err != nil {
			return loc, err
		}
		loc.PkgPath = m.DebugLoc.PkgPath
		loc.File = path.Clean(loc.File)
		loc.Line = line
		return loc, nil
	}
	if strings.HasPrefix(arg, "+") || strings.HasPrefix(arg, "-") {
		// Location is specified as offset from current file.
		if line, err = strconv.Atoi(arg); err != nil {
			return
		}
		loc.Line += line
		return loc, nil
	}
	if line, err = strconv.Atoi(arg); err == nil {
		// Location is the line number in the current file.
		loc.Line = line
		return loc, nil
	}
	return
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

func debugContinue(m *Machine, arg string) error { m.DebugState = DebugAtRun; return nil }

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
const listUsage = `list|l`
const listShort = `Show source code.`

func debugList(m *Machine, arg string) error {
	debugLineInfo(m)
	lines, offset, err := sourceLines(m.DebugLoc.File, m.DebugLoc.Line)
	if err != nil {
		return err
	}
	for i, line := range lines {
		cursor := ""
		if m.DebugLoc.Line == i+offset {
			cursor = "=>"
		}
		fmt.Fprintf(m.DebugOut, "%2s %4d: %s\n", cursor, i+offset, line)
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
	println("not implemented yet")
	return nil
}
