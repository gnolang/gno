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
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
)

// DebugState is the state of the machine debugger, defined by a finite state
// automaton with the following transitions, evaluated at each debugger input
// or each gnoVM instruction while in step mode:
// - DebugAtInit -> DebugAtCmd: initial debugger setup is performed
// - DebugAtCmd  -> DebugAtCmd: when command is for inspecting or setting a breakpoint
// - DebugAtCmd  -> DebuAtRun:  when command is 'continue', 'step' or 'stepi'
// - DebugAtCmd  -> DebugAtExit: when command is 'quit' or 'resume'
// - DebugAtRun  -> DebugAtRun: when current machine instruction doesn't match a breakpoint
// - DebugAtRun  -> DebugAtCmd: when current machine instruction matches a breakpoint
// - DebugAtRun  -> DebugAtExit: when the program terminates
type DebugState int

const (
	DebugAtInit DebugState = iota // performs debugger IO setup and enters gnoVM in step mode
	DebugAtCmd                    // awaits a new command from the debugger input stream
	DebugAtRun                    // awaits the next machine instruction
	DebugAtExit                   // closes debugger IO and exits gnoVM from step mode
)

// Debugger describes a machine debugger.
type Debugger struct {
	enabled bool           // when true, machine is in step mode
	in      io.Reader      // debugger input, defaults to Stdin
	out     io.Writer      // debugger output, defaults to Stdout
	scanner *bufio.Scanner // to parse input per line

	state       DebugState                  // current state of debugger
	lastCmd     string                      // last debugger command
	lastArg     string                      // last debugger command arguments
	loc         Location                    // source location of the current machine instruction
	prevLoc     Location                    // source location of the previous machine instruction
	nextLoc     Location                    // source location at the 'next' command
	breakpoints []Location                  // list of breakpoints set by user, as source locations
	call        []Location                  // for function tracking, ideally should be provided by machine frame
	frameLevel  int                         // frame level of the current machine instruction
	nextDepth   int                         // function call depth at the 'next' command
	getSrc      func(string, string) string // helper to access source from repl or others
	rootDir     string
}

// Enable makes the debugger d active, using in as input reader, out as output writer and f as a source helper.
func (d *Debugger) Enable(in io.Reader, out io.Writer, f func(string, string) string) {
	d.in = in
	d.out = out
	d.enabled = true
	d.state = DebugAtInit
	d.getSrc = f
	d.rootDir = gnoenv.RootDir()
}

// Disable makes the debugger d inactive.
func (d *Debugger) Disable() {
	d.enabled = false
	d.loc = Location{}
	d.prevLoc = Location{}
	d.nextLoc = Location{}
}

type debugCommand struct {
	debugFunc          func(*Machine, string) error // debug command
	usage, short, long string                       // command help texts
}

var (
	debugCmds     map[string]debugCommand
	debugCmdNames []string
)

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
		"next":        {debugContinue, nextUsage, nextShort, ""},
		"step":        {debugContinue, stepUsage, stepShort, ""},
		"stepi":       {debugContinue, stepiUsage, stepiShort, ""},
		"stepout":     {debugContinue, stepoutUsage, stepoutShort, ""},
		"up":          {debugUp, upUsage, upShort, ""},
	}

	// Sort command names for help.
	debugCmdNames = make([]string, 0, len(debugCmds))
	for name := range debugCmds {
		debugCmdNames = append(debugCmdNames, name)
	}
	sort.Strings(debugCmdNames)

	// Set command aliases.
	debugCmds["b"] = debugCmds["break"]
	debugCmds["bp"] = debugCmds["breakpoints"]
	debugCmds["bt"] = debugCmds["stack"]
	debugCmds["c"] = debugCmds["continue"]
	debugCmds["h"] = debugCmds["help"]
	debugCmds["l"] = debugCmds["list"]
	debugCmds["n"] = debugCmds["next"]
	debugCmds["p"] = debugCmds["print"]
	debugCmds["quit"] = debugCmds["exit"]
	debugCmds["q"] = debugCmds["exit"]
	debugCmds["s"] = debugCmds["step"]
	debugCmds["si"] = debugCmds["stepi"]
	debugCmds["so"] = debugCmds["stepout"]
}

// Debug is the debug callback invoked at each VM execution step. It implements the DebugState FSA.
func (m *Machine) Debug() {
loop:
	for {
		switch m.Debugger.state {
		case DebugAtInit:
			debugUpdateLocation(m)
			fmt.Fprintln(m.Debugger.out, "Welcome to the Gnovm debugger. Type 'help' for list of commands.")
			m.Debugger.scanner = bufio.NewScanner(m.Debugger.in)
			m.Debugger.state = DebugAtCmd
		case DebugAtCmd:
			if err := debugCmd(m); err != nil {
				fmt.Fprintln(m.Debugger.out, "Command failed:", err)
			}
		case DebugAtRun:
			if !m.Debugger.enabled {
				break loop
			}
			switch m.Debugger.lastCmd {
			case "si", "stepi":
				m.Debugger.state = DebugAtCmd
				debugLineInfo(m)
			case "s", "step":
				if m.Debugger.loc != m.Debugger.prevLoc && m.Debugger.loc.File != "" {
					m.Debugger.state = DebugAtCmd
					m.Debugger.prevLoc = m.Debugger.loc
					debugList(m, "")
					continue loop
				}
			case "n", "next":
				if m.Debugger.loc != m.Debugger.prevLoc && m.Debugger.loc.File != "" &&
					(m.Debugger.nextDepth == 0 || !sameLine(m.Debugger.loc, m.Debugger.nextLoc) && callDepth(m) <= m.Debugger.nextDepth) {
					m.Debugger.state = DebugAtCmd
					m.Debugger.prevLoc = m.Debugger.loc
					debugList(m, "")
					continue loop
				}
			case "stepout", "so":
				if callDepth(m) < m.Debugger.nextDepth {
					m.Debugger.state = DebugAtCmd
					m.Debugger.prevLoc = m.Debugger.loc
					debugList(m, "")
					continue loop
				}
			default:
				if atBreak(m) {
					m.Debugger.state = DebugAtCmd
					m.Debugger.prevLoc = m.Debugger.loc
					debugList(m, "")
					continue loop
				}
			}
			break loop
		case DebugAtExit:
			os.Exit(0)
		}
	}
	m.Debugger.prevLoc = m.Debugger.loc
	debugUpdateLocation(m)

	// Keep track of exact locations when performing calls.
	op := m.Ops[len(m.Ops)-1]
	switch op {
	case OpCall:
		m.Debugger.call = append(m.Debugger.call, m.Debugger.loc)
	case OpReturn, OpReturnFromBlock:
		m.Debugger.call = m.Debugger.call[:len(m.Debugger.call)-1]
	}
}

// callDepth returns the function call depth.
func callDepth(m *Machine) int {
	n := 0
	for _, f := range m.Frames {
		if f.Func == nil {
			continue
		}
		n++
	}
	return n
}

// sameLine returns true if both arguments are at the same line.
func sameLine(loc1, loc2 Location) bool {
	return loc1.PkgPath == loc2.PkgPath && loc1.File == loc2.File && loc1.Line == loc2.Line
}

// atBreak returns true if current machine location matches a breakpoint, false otherwise.
func atBreak(m *Machine) bool {
	loc := m.Debugger.loc
	if loc == m.Debugger.prevLoc {
		return false
	}
	for _, b := range m.Debugger.breakpoints {
		if loc.File == b.File && loc.Line == b.Line {
			return true
		}
	}
	return false
}

// debugCmd processes a debugger REPL command. It displays a prompt, then
// reads and parses a command from the debugger input stream, then executes
// the corresponding function or returns an error.
// If the command is empty, the last non-empty command is repeated.
func debugCmd(m *Machine) error {
	var cmd, arg string
	fmt.Fprint(m.Debugger.out, "dbg> ")
	if !m.Debugger.scanner.Scan() {
		return debugDetach(m, arg) // Clean close of debugger, the target program resumes.
	}
	line := trimLeftSpace(m.Debugger.scanner.Text())
	if i := indexSpace(line); i >= 0 {
		cmd = line[:i]
		arg = trimLeftSpace(line[i:])
	} else {
		cmd = line
	}
	if cmd == "" {
		if m.Debugger.lastCmd == "" {
			return nil
		}
		cmd, arg = m.Debugger.lastCmd, m.Debugger.lastArg
	} else if cmd[0] == '#' {
		return nil
	}
	c, ok := debugCmds[cmd]
	if !ok {
		return errors.New("command not available: " + cmd)
	}
	m.Debugger.lastCmd, m.Debugger.lastArg = cmd, arg
	return c.debugFunc(m, arg)
}

func trimLeftSpace(s string) string { return strings.TrimLeftFunc(s, unicode.IsSpace) }
func indexSpace(s string) int       { return strings.IndexFunc(s, unicode.IsSpace) }

// Serve waits for a remote client to connect to addr and use this connection for debugger IO.
// It returns an error if the connection can not be established, or nil.
func (d *Debugger) Serve(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	print("Waiting for debugger client to connect at ", addr)
	conn, err := l.Accept()
	if err != nil {
		return err
	}
	println(" connected!")
	d.in, d.out = conn, conn
	return nil
}

// debugUpdateLocation computes the source code location for the current VM state.
// The result is stored in Debugger.DebugLoc.
func debugUpdateLocation(m *Machine) {
	loc := m.LastBlock().GetSource(m.Store).GetLocation()

	if loc.PkgPath == "repl" {
		loc.File = "<repl>"
	}

	if m.Debugger.loc.PkgPath == "" ||
		loc.PkgPath != "" && loc.PkgPath != m.Debugger.loc.PkgPath ||
		loc.File != "" && loc.File != m.Debugger.loc.File {
		m.Debugger.loc = loc
	}

	// The location computed from above points to the block start. Examine
	// expressions and statements to have the exact line within the block.

	nx := len(m.Exprs)
	for i := nx - 1; i >= 0; i-- {
		expr := m.Exprs[i]
		if l := expr.GetLine(); l > 0 {
			if col := expr.GetColumn(); col > 0 {
				m.Debugger.loc.Line = l
				m.Debugger.loc.Column = expr.GetColumn()
			}
			return
		}
	}

	if len(m.Stmts) > 0 {
		if stmt := m.PeekStmt1(); stmt != nil {
			if l := stmt.GetLine(); l > 0 {
				if col := stmt.GetColumn(); col > 0 {
					m.Debugger.loc.Line = l
					m.Debugger.loc.Column = stmt.GetColumn()
				}
				return
			}
		}
	}
}

// ---------------------------------------
const (
	breakUsage = `break|b [locspec]`
	breakShort = `Set a breakpoint.`
	breakLong  = `
The syntax accepted for locspec is:
- <filename>:<line> specifies the line in filename. Filename can be relative.
- <line> specifies the line in the current source file.
- +<offset> specifies the line offset lines after the current one.
- -<offset> specifies the line offset lines before the current one.
`
)

func debugBreak(m *Machine, arg string) error {
	loc, err := parseLocSpec(m, arg)
	if err != nil {
		return err
	}
	m.Debugger.breakpoints = append(m.Debugger.breakpoints, loc)
	printBreakpoint(m, len(m.Debugger.breakpoints)-1)
	return nil
}

func printBreakpoint(m *Machine, i int) {
	b := m.Debugger.breakpoints[i]
	fmt.Fprintf(m.Debugger.out, "Breakpoint %d at %s %s\n", i, b.PkgPath, b)
}

func parseLocSpec(m *Machine, arg string) (loc Location, err error) {
	var line int
	loc = m.Debugger.loc

	if strings.Contains(arg, ":") {
		// Location is specified by filename:line.
		strs := strings.Split(arg, ":")
		if strs[0] != "" {
			if loc.File, err = filepath.Abs(strs[0]); err != nil {
				return loc, err
			}
			loc.File = path.Clean(loc.File)
			if m.Debugger.rootDir != "" && strings.HasPrefix(loc.File, m.Debugger.rootDir) {
				loc.File = strings.TrimPrefix(loc.File, m.Debugger.rootDir+"/gnovm/stdlibs/")
				loc.File = strings.TrimPrefix(loc.File, m.Debugger.rootDir+"/examples/")
				loc.File = strings.TrimPrefix(loc.File, m.Debugger.rootDir+"/")
				loc.PkgPath = path.Dir(loc.File)
				loc.File = path.Base(loc.File)
			}
		}
		if line, err = strconv.Atoi(strs[1]); err != nil {
			return loc, err
		}
		loc.Line = line
		return loc, nil
	}
	// Location is in the current file.
	if loc.File == "" {
		return loc, errors.New("unknown source file")
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
const (
	breakpointsUsage = `breakpoints|bp`
	breakpointsShort = `Print out info for active breakpoints.`
)

func debugBreakpoints(m *Machine, arg string) error {
	for i := range m.Debugger.breakpoints {
		printBreakpoint(m, i)
	}
	return nil
}

// ---------------------------------------
const (
	clearUsage = `clear [id]`
	clearShort = `Delete breakpoint (all if no id).`
)

func debugClear(m *Machine, arg string) error {
	if arg != "" {
		id, err := strconv.Atoi(arg)
		if err != nil || id < 0 || id >= len(m.Debugger.breakpoints) {
			return fmt.Errorf("invalid breakpoint id: %v", arg)
		}
		m.Debugger.breakpoints = slices.Delete(m.Debugger.breakpoints, id, id+1)
		return nil
	}
	m.Debugger.breakpoints = nil
	return nil
}

// ---------------------------------------
// NOTE: the difference between continue, next, step, stepi and stepout is handled within the Debug() loop.
const (
	continueUsage = `continue|c`
	continueShort = `Run until breakpoint or program termination.`

	nextUsage = `next|n`
	nextShort = `Step over to next source line.`

	stepUsage = `step|s`
	stepShort = `Single step through program.`

	stepiUsage = `stepi|si`
	stepiShort = `Single step a single VM instruction.`

	stepoutUsage = `stepout|so`
	stepoutShort = `Step out of the current function.`
)

func debugContinue(m *Machine, arg string) error {
	m.Debugger.state = DebugAtRun
	m.Debugger.frameLevel = 0
	m.Debugger.nextDepth = callDepth(m)
	m.Debugger.nextLoc = m.Debugger.loc
	return nil
}

// ---------------------------------------
const (
	detachUsage = `detach`
	detachShort = `Close debugger and resume program.`
)

func debugDetach(m *Machine, arg string) error {
	m.Debugger.enabled = false
	m.Debugger.state = DebugAtRun
	if i, ok := m.Debugger.in.(io.Closer); ok {
		i.Close()
	}
	return nil
}

// ---------------------------------------
const (
	downUsage = `down [n]`
	downShort = `Move the current frame down by n (default 1).`
)

func debugDown(m *Machine, arg string) (err error) {
	n := 1
	if arg != "" {
		if n, err = strconv.Atoi(arg); err != nil {
			return err
		}
	}
	if level := m.Debugger.frameLevel - n; level >= 0 && level < len(m.Debugger.call) {
		m.Debugger.frameLevel = level
	}
	debugList(m, "")
	return nil
}

// ---------------------------------------
const (
	exitUsage = `exit|quit|q`
	exitShort = `Exit the debugger and program.`
)

func debugExit(m *Machine, arg string) error { m.Debugger.state = DebugAtExit; return nil }

// ---------------------------------------
const (
	helpUsage = `help|h [command]`
	helpShort = `Print the help message.`
)

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
		fmt.Fprintln(m.Debugger.out, t)
		return nil
	}
	t := "The following commands are available:\n\n"
	for _, name := range debugCmdNames {
		c := debugCmds[name]
		t += fmt.Sprintf("%-25s %s\n", c.usage, c.short)
	}
	t += "\nType help followed by a command for full documentation."
	fmt.Fprintln(m.Debugger.out, t)
	return nil
}

// ---------------------------------------
const (
	listUsage = `list|l [locspec]`
	listShort = `Show source code.`
	listLong  = `
See 'help break' for locspec syntax. If locspec is empty,
list shows the source code around the current line.
`
)

func debugList(m *Machine, arg string) (err error) {
	loc := m.Debugger.loc
	hideCursor := false

	if arg == "" {
		debugLineInfo(m)
		if m.Debugger.lastCmd == "up" || m.Debugger.lastCmd == "down" {
			loc = debugFrameLoc(m, m.Debugger.frameLevel)
			fmt.Fprintf(m.Debugger.out, "Frame %d: %s\n", m.Debugger.frameLevel, loc)
		}
	} else {
		if loc, err = parseLocSpec(m, arg); err != nil {
			return err
		}
		hideCursor = true
		fmt.Fprintf(m.Debugger.out, "Showing %s\n", loc)
	}
	if loc.File == "" && (m.Debugger.lastCmd == "list" || m.Debugger.lastCmd == "l") {
		return errors.New("unknown source file")
	}
	src, err := fileContent(m.Store, loc.PkgPath, loc.File)
	if err != nil {
		// Use optional getSrc helper as fallback to get source.
		if m.Debugger.getSrc != nil {
			src = m.Debugger.getSrc(loc.PkgPath, loc.File)
		}
		if src == "" {
			return err
		}
	}
	lines, offset := linesAround(src, loc.Line, 10)
	for i, l := range lines {
		cursor := ""
		if !hideCursor && loc.Line == i+offset {
			cursor = "=>"
		}
		fmt.Fprintf(m.Debugger.out, "%2s %4d: %s\n", cursor, i+offset, l)
	}
	return nil
}

func debugLineInfo(m *Machine) {
	if m.Debugger.loc.File == "" {
		return
	}
	line := string(m.Package.PkgName)
	if len(m.Frames) > 0 {
		f := m.Frames[len(m.Frames)-1]
		if f.Func != nil {
			line += "." + string(f.Func.Name) + "()"
		}
	}
	fmt.Fprintf(m.Debugger.out, "> %s %s\n", line, m.Debugger.loc)
}

func isMemPackage(st Store, pkgPath string) bool {
	ds, ok := st.(*defaultStore)
	return ok && ds.iavlStore.Has([]byte(backendPackagePathKey(pkgPath)))
}

func fileContent(st Store, pkgPath, name string) (string, error) {
	if isMemPackage(st, pkgPath) {
		return st.GetMemFile(pkgPath, name).Body, nil
	}
	buf, err := os.ReadFile(name)
	return string(buf), err
}

func linesAround(src string, index, n int) ([]string, int) {
	lines := strings.Split(src, "\n")
	start := max(1, index-n/2) - 1
	end := min(start+n, len(lines))
	if start >= end {
		start = max(1, end-n)
	}
	return lines[start:end], start + 1
}

// ---------------------------------------
const (
	printUsage = `print|p <expression>`
	printShort = `Print a variable or expression.`
)

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
	fmt.Fprintln(m.Debugger.out, tv)
	return nil
}

// debugEvalExpr evaluates a Go expression in the context of the VM and returns
// the corresponding typed value, or an error.
// The supported expression syntax is a small subset of Go expressions:
// basic literals, identifiers, selectors, index expressions, or a combination
// of those are supported, but none of function calls, arithmetic, logic or
// assign operations, type assertions of convertions.
// This is sufficient for a debugger to perform 'print (*f).S[x][y]' for example.
func debugEvalExpr(m *Machine, node ast.Node) (tv TypedValue, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

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
			r, _, _, err := strconv.UnquoteChar(n.Value[1:len(n.Value)-1], 0)
			if err != nil {
				return tv, err
			}
			return typedRune(r), nil
		case token.STRING:
			s, err := strconv.Unquote(n.Value)
			if err != nil {
				return tv, err
			}
			return typedString(s), nil
		}
		return tv, fmt.Errorf("invalid basic literal value: %s", n.Value)
	case *ast.Ident:
		if tv, ok := debugLookup(m, n.Name); ok {
			return tv, nil
		}
		// Rewritten loopvar.
		if tv, ok := debugLookup(m, fmt.Sprintf("%s%s", ".loopvar_", n.Name)); ok {
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
		if pv, ok := x.V.(*PackageValue); ok {
			if i, ok := pv.Block.(*Block).Source.GetLocalIndex(Name(n.Sel.Name)); ok {
				return pv.Block.(*Block).Values[i], nil
			}
			return tv, fmt.Errorf("invalid selector: %s", n.Sel.Name)
		}
		tr, _, _, _, _ := findEmbeddedFieldType(x.T.GetPkgPath(), x.T, Name(n.Sel.Name), nil)
		if len(tr) == 0 {
			return tv, fmt.Errorf("invalid selector: %s", n.Sel.Name)
		}
		for _, vp := range tr {
			x = x.GetPointerToFromTV(m.Alloc, m.Store, vp).Deref()
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
		return x.GetPointerAtIndex(m.Realm, m.Alloc, m.Store, &index).Deref(), nil
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
		if ncall == m.Debugger.frameLevel {
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

	// XXX The following logic isn't necessary and it isn't correct either.
	// XXX See `GetPathForName(Store, Name) ValuePath` in node.go,
	// XXX get the path and pass it into the last block.GetPointerTo().
	// XXX That function will find the correct block by depth etc.
	// XXX There was some latent bug for case:
	// XXX '{in: "b 37\nc\np b\n", out: "(3 int)"},' (debugger test case #51)
	// XXX which was revealed by some earlier commits regarding lines
	// XXX (Node now has not just the starting .Pos but also .End.)
	// XXX and is resolved by the following diff to values.go:
	// XXX The exact bug probably doesn't matter, as the logic
	// XXX should be replaced by the aforementioned block.GetPointerTo().
	//
	// --- a/gnovm/pkg/gnolang/values.go
	// +++ b/gnovm/pkg/gnolang/values.go
	// @@ -2480,6 +2480,7 @@ func (b *Block) ExpandWith(alloc *Allocator, source BlockNode) {
	//                 }
	//         }
	//         b.Values = values
	// +       b.Source = source // otherwise new variables won't show in print or debugger.
	//  }

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

	// Search value in current frame level blocks, or main scope.
	for _, b := range sblocks {
		switch t := b.Source.(type) {
		case *IfStmt:
			for i, s := range ifBody(m, t).Source.GetBlockNames() {
				if string(s) == name {
					return b.Values[i], true
				}
			}
		}
		for i, s := range b.Source.GetBlockNames() {
			if string(s) == name {
				return b.Values[i], true
			}
		}
	}
	// Fallback: search a global value.
	if v := sblocks[0].Source.GetSlot(m.Store, Name(name), true); v != nil {
		return *v, true
	}
	return tv, false
}

// ifBody returns the Then or Else body corresponding to the current location.
func ifBody(m *Machine, ifStmt *IfStmt) IfCaseStmt {
	if l := ifStmt.Else.GetLocation().Line; l > 0 && debugFrameLoc(m, m.Debugger.frameLevel).Line > l {
		return ifStmt.Else
	}
	return ifStmt.Then
}

// ---------------------------------------
const (
	stackUsage = `stack|bt`
	stackShort = `Print stack trace.`
)

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
		fmt.Fprintf(m.Debugger.out, "%d\tin %s\n\tat %s\n", i, fname, loc)
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
	if n == 0 || len(m.Debugger.call) == 0 {
		return m.Debugger.loc
	}
	return m.Debugger.call[len(m.Debugger.call)-n]
}

// ---------------------------------------
const (
	upUsage = `up [n]`
	upShort = `Move the current frame up by n (default 1).`
)

func debugUp(m *Machine, arg string) (err error) {
	n := 1
	if arg != "" {
		if n, err = strconv.Atoi(arg); err != nil {
			return err
		}
	}
	if level := m.Debugger.frameLevel + n; level >= 0 && level < len(m.Debugger.call) {
		m.Debugger.frameLevel = level
	}
	debugList(m, "")
	return nil
}
