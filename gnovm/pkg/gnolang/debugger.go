package gnolang

import (
	"bufio"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
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
	DebugAddr    string    // optional address [host]:port for DebugIn/DebugOut
	DebugIn      io.Reader // debugger input, defaults to Stdin
	DebugOut     io.Writer // debugger output, defaults to Stdout
	DebugScanner *bufio.Scanner

	DebugState
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
	debugCmdNames = sort.Strings(maps.Keys(debugCmds))

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
			debugUpdateLocation(m)
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
	debugUpdateLocation(m)

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
	if !m.DebugScanner.Scan() {
		return debugDetach(m, arg) // Clean close of debugger, the target program resumes.
	}
	line := m.DebugScanner.Text()
	n, _ := fmt.Sscan(line, &cmd, &arg)
	if n == 0 {
		if m.lastDebugCmd == "" {
			return nil
		}
		cmd, arg = m.lastDebugCmd, m.lastDebugArg
	} else if cmd[0] == '#' {
		return nil
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
	if m.DebugAddr != "" {
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
	m.DebugScanner = bufio.NewScanner(m.DebugIn)
}

// debugUpdateLocation computes the source code location for the current VM state.
// The result is stored in Debugger.DebugLoc.
func debugUpdateLocation(m *Machine) {
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
	if i, ok := m.DebugIn.(io.Closer); ok {
		i.Close()
	}
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
	t := "The following commands are available:\n\n"
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

func debugPrint(m *Machine, arg string) (err error) {
	if arg == "" {
		return errors.New("missing argument")
	}
	// Use the Go parser to get the AST representation of print argument as a Go expresssion.
	ast, err := parser.ParseExpr(arg)
	if err != nil {
		return err
	}
	tv, err := debugEvalExpr(m, ast)
	if err != nil {
		return err
	}
	fmt.Println(m.DebugOut, tv)
	return nil
}

// debugEvalExpr evaluates a Go expression in the context of the VM and returns
// the corresponding value, or an error.
// The supported expression syntax is a small subset of Go expressions:
// basic literals, identifiers, selectors, index expressions, or a combination
// of those are supported, but none of function calls, arithmetic, logic or
// assign operations, type assertions of convertions.
// This is sufficient for a debugger to perform 'print (*f).S[x][y]' for example.
func debugEvalExpr(m *Machine, node ast.Node) (tv TypedValue, err error) {
	switch n := node.(type) {
	case *ast.BasicLit:
		switch n.Kind {
		case token.INT:
			i, err := strconv.ParseInt(n.Value, 0, 0)
			if err != nil {
				return tv, err
			}
			return typedInt(int(i)), nil
		case token.CHAR:
			return typedRune(([]rune(n.Value))[0]), nil
		case token.STRING:
			return typedString(n.Value), nil
		}
		return tv, fmt.Errorf("invalid basic literal value: %s", n.Value)
	case *ast.Ident:
		if tv, ok := debugLookup(m, n.Name); ok {
			return tv, nil
		}
		return tv, fmt.Errorf("could not find symbol value for %s", n.Name)
	case *ast.ParenExpr:
		return debugEvalExpr(m, n.X)
	case *ast.StarExpr:
		x, err := debugEvalExpr(m, n.X)
		if err != nil {
			return tv, err
		}
		pv, ok := x.V.(PointerValue)
		if !ok {
			return tv, fmt.Errorf("Not a pointer value: %v", x)
		}
		return pv.Deref(), nil
	case *ast.SelectorExpr:
		x, err := debugEvalExpr(m, n.X)
		if err != nil {
			return tv, err
		}
		// TODO: handle selector on package.
		tr, _, _, _, _ := findEmbeddedFieldType(x.T.GetPkgPath(), x.T, Name(n.Sel.Name), nil)
		if len(tr) == 0 {
			return tv, fmt.Errorf("invalid selector: %s", n.Sel.Name)
		}
		for _, vp := range tr {
			x = x.GetPointerTo(m.Alloc, m.Store, vp).Deref()
		}
		return x, nil
	case *ast.IndexExpr:
		x, err := debugEvalExpr(m, n.X)
		if err != nil {
			return tv, err
		}
		index, err := debugEvalExpr(m, n.Index)
		if err != nil {
			return tv, err
		}
		return x.GetPointerAtIndex(m.Alloc, m.Store, &index).Deref(), nil
	default:
		err = fmt.Errorf("expression not supported: %v", n)
	}
	return tv, err
}

// debugLookup returns the current VM value corresponding to name ident in
// the current function call frame, or the global frame if not found.
// Note: the commands 'up' and 'down' change the frame level to start from.
func debugLookup(m *Machine, name string) (tv TypedValue, ok bool) {
	// Position to the right frame.
	ncall := 0
	var i int
	var fblocks []BlockNode
	var funBlock BlockNode
	for i = len(m.Frames) - 1; i >= 0; i-- {
		if m.Frames[i].Func != nil {
			funBlock = m.Frames[i].Func.Source
		}
		if ncall == m.debugFrameLevel {
			break
		}
		if m.Frames[i].Func != nil {
			fblocks = append(fblocks, m.Frames[i].Func.Source)
			ncall++
		}
	}
	if i < 0 {
		return tv, false
	}

	// Position to the right block, i.e the first after the last fblock (if any).
	for i = len(m.Blocks) - 1; i >= 0; i-- {
		if len(fblocks) == 0 {
			break
		}
		if m.Blocks[i].Source == fblocks[0] {
			fblocks = fblocks[1:]
		}
	}
	if i < 0 {
		return tv, false
	}

	// get SourceBlocks in the same frame level.
	var sblocks []*Block
	for ; i >= 0; i-- {
		sblocks = append(sblocks, m.Blocks[i])
		if m.Blocks[i].Source == funBlock {
			break
		}
	}
	if i > 0 {
		sblocks = append(sblocks, m.Blocks[0]) // Add global block
	}

	// Search value in current frame level blocks, or main.
	for _, b := range sblocks {
		for i, s := range b.Source.GetBlockNames() {
			if string(s) == name {
				return b.Values[i], true
			}
		}
	}
	return tv, false
}

// ---------------------------------------
const stackUsage = `stack|bt`
const stackShort = `Print stack trace.`

func debugStack(m *Machine, arg string) error {
	i := 0
	for {
		ff := debugFrameFunc(m, i)
		loc := debugFrameLoc(m, i)
		if ff == nil {
			break
		}
		var fname string
		if ff.IsMethod {
			fname = fmt.Sprintf("%v.(%v).%v", ff.PkgPath, ff.Type.(*FuncType).Params[0].Type, ff.Name)
		} else {
			fname = fmt.Sprintf("%v.%v", ff.PkgPath, ff.Name)
		}
		fmt.Fprintf(m.DebugOut, "%d\tin %s\n\tat %s:%d\n", i, fname, loc.File, loc.Line)
		i++
	}
	return nil
}

func debugFrameFunc(m *Machine, n int) *FuncValue {
	for ncall, i := 0, len(m.Frames)-1; i >= 0; i-- {
		f := m.Frames[i]
		if f.Func == nil {
			continue
		}
		if ncall == n {
			return f.Func
		}
		ncall++
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
