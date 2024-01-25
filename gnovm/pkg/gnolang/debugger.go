package gnolang

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sort"
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
	DebugLoc     Location
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
		"continue": {debugContinue, continueUsage, continueShort, ""},
		"detach":   {debugDetach, detachUsage, detachShort, ""},
		"exit":     {debugExit, exitUsage, exitShort, ""},
		"help":     {debugHelp, helpUsage, helpShort, ""},
		"print":    {debugPrint, printUsage, printShort, ""},
		"stepi":    {debugStepi, stepiUsage, stepiShort, ""},
	}

	// Sort command names for help.
	debugCmdNames = make([]string, 0, len(debugCmds))
	for name := range debugCmds {
		debugCmdNames = append(debugCmdNames, name)
	}
	sort.SliceStable(debugCmdNames, func(i, j int) bool { return debugCmdNames[i] < debugCmdNames[j] })

	// Set command aliases.
	debugCmds["c"] = debugCmds["continue"]
	debugCmds["h"] = debugCmds["help"]
	debugCmds["p"] = debugCmds["print"]
	debugCmds["quit"] = debugCmds["exit"]
	debugCmds["q"] = debugCmds["exit"]
	debugCmds["si"] = debugCmds["stepi"]
}

// Debug is the debug callback invoked at each VM execution step.
func (m *Machine) Debug() {
loop:
	for {
		switch m.DebugState {
		case DebugAtInit:
			initDebugIO(m)
			fmt.Fprintln(m.DebugOut, "Welcome to the Gnovm debugger. Type 'help' for list of commands.")
			m.DebugState = DebugAtCmd
		case DebugAtCmd:
			if err := debugCmd(m); err != nil {
				fmt.Fprintln(m.DebugOut, "Command failed:", err)
			}
		case DebugAtRun:
			// TODO: here, check matching breakpoint condition and set DebugAtCmd if match.
			if m.lastDebugCmd == "stepi" || m.lastDebugCmd == "si" {
				m.DebugState = DebugAtCmd
			}
			break loop
		case DebugAtExit:
			os.Exit(0)
		}
	}
	debugUpdateLoc(m)
	fmt.Fprintln(m.DebugOut, "in debug:", m.DebugLoc)
}

// debugCmd processes a debugger REPL command. It displays a prompt, then
// reads and parses a command from the debugger input stream, then executes
// the corresponding function or returns an error.
func debugCmd(m *Machine) error {
	var cmd, arg string
	fmt.Fprint(m.DebugOut, "dbg> ")
	if n, err := fmt.Fscanln(m.DebugIn, &cmd, &arg); errors.Is(err, io.EOF) {
		return debugDetach(m, arg) // Clean close of debugger, the target program resumes.
	} else if n == 0 {
		return nil
	}
	c, ok := debugCmds[cmd]
	if !ok {
		return errors.New("command not available: " + cmd)
	}
	m.lastDebugCmd = cmd
	return c.debugFunc(m, arg)
}

// initDebugIO initializes the debugger standard input and output streams.
// If no debug address was specified at program start, the debugger will inherit its
// standard input and output from the process, which will be shared with the target program.
// If the debug address is specified, the program will be blocked until
// a client connection is established. The debugger will use this connection and
// not affecting the target program's.
// An error at connection setting will result in program panic.
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

// debugUpdateLoc computes the source code location of the last VM operation.
// The result is stored in Debugger.DebugLoc.
func debugUpdateLoc(m *Machine) {
	loc := m.LastBlock().Source.GetLocation()

	if m.DebugLoc.PkgPath == "" ||
		loc.PkgPath != "" && loc.PkgPath != m.DebugLoc.PkgPath ||
		loc.File != "" && loc.File != m.DebugLoc.File {
		m.Debugger.DebugLoc = loc
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
const continueUsage = `continue|c`
const continueShort = `Run until breakpoint or program termination.`

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
			t += "\n\n" + c.long
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
const printUsage = `print|p <expression>`
const printShort = `Print a variable or expression.`

func debugPrint(m *Machine, arg string) error {
	println("not implemented yet")
	return nil
}

// ---------------------------------------
const stepiUsage = `stepi|si`
const stepiShort = `Single step a single VM instruction.`

func debugStepi(m *Machine, arg string) error { m.DebugState = DebugAtRun; return nil }
