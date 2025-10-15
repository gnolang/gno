package gnolang

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// DAPMode represents the DAP server connection mode
type DAPMode int

const (
	_ DAPMode = iota
	DAPModeStdio
	DAPModeTCP
)

// DAPServer implements the Debug Adapter Protocol server
type DAPServer struct {
	debugger *Debugger
	machine  *Machine

	// Connection handling
	mode   DAPMode
	conn   net.Conn
	reader *bufio.Reader
	writer io.Writer

	// Protocol state
	seq         int32
	mu          sync.Mutex
	initialized bool
	terminated  bool

	// Debugging state
	breakpoints      map[string][]Breakpoint // map[source.path][]Breakpoint
	nextBreakpointID int
	threadID         int         // Gno VM is single-threaded, but DAP requires thread IDs
	attachMode       bool        // true if in attach mode (program not auto-started)
	programFiles     []*FileNode // files to run when attach completes
	mainExpr         string      // main expression to evaluate

	// Variable reference management
	variableRefs    map[int]variableInfo // map[reference]info
	nextVariableRef int

	// Client capabilities
	clientLinesStartAt1   bool
	clientColumnsStartAt1 bool
	clientPathFormat      string

	// Channels for communication
	stopCh     chan StopReason
	continueCh chan struct{}
}

// StopReason represents why execution stopped
type StopReason struct {
	Reason      string
	Description string
}

// variableInfo stores information about a variable with children
type variableInfo struct {
	value    TypedValue
	frameID  int
	isScoped bool // true if this is a scope reference (locals/globals)
}

// dapOutputWriter intercepts output and sends it as DAP output events
type dapOutputWriter struct {
	server   *DAPServer
	category string
}

func (w *dapOutputWriter) Write(p []byte) (n int, err error) {
	// Send output event if DAP server is connected
	if w.server != nil && w.server.writer != nil {
		event := &OutputEvent{
			Event: *NewEvent("output"),
			Body: OutputEventBody{
				Category: w.category,
				Output:   string(p),
			},
		}
		w.server.sendMessage(event)
	}
	return len(p), nil
}

// NewDAPServer creates a new DAP server
func NewDAPServer(debugger *Debugger, machine *Machine) *DAPServer {
	return &DAPServer{
		debugger:              debugger,
		machine:               machine,
		breakpoints:           make(map[string][]Breakpoint),
		variableRefs:          make(map[int]variableInfo),
		nextVariableRef:       3000, // Start from 3000 to avoid conflicts with scope references
		threadID:              1,    // Single thread
		stopCh:                make(chan StopReason, 1),
		continueCh:            make(chan struct{}, 1),
		clientLinesStartAt1:   true,
		clientColumnsStartAt1: true,
	}
}

// Serve starts the DAP server
func (s *DAPServer) Serve(mode DAPMode, addr string) error {
	s.mode = mode

	switch mode {
	case DAPModeStdio:
		s.reader = bufio.NewReader(s.debugger.in)
		s.writer = s.debugger.out
	case DAPModeTCP:
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to listen on %s: %w", addr, err)
		}
		defer listener.Close()

		fmt.Printf("DAP server listening on %s\n", addr)
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept connection: %w", err)
		}
		s.conn = conn
		s.reader = bufio.NewReader(conn)
		s.writer = conn
		defer conn.Close()
	}

	// Set up output redirection to DAP
	if s.machine != nil {
		s.machine.Output = &dapOutputWriter{
			server:   s,
			category: "stdout",
		}
	}

	// Start the main message loop
	return s.messageLoop()
}

// messageLoop processes incoming DAP messages
func (s *DAPServer) messageLoop() error {
	for !s.terminated {
		msg, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("failed to read message: %w", err)
		}

		if err := s.handleMessage(msg); err != nil {
			// Log error but continue processing
			s.sendErrorResponse(msg, err.Error())
		}
	}
	return nil
}

// readMessage reads a DAP message from the input
func (s *DAPServer) readMessage() ([]byte, error) {
	// DAP uses HTTP-like headers: "Content-Length: <length>\r\n\r\n<json>"
	headers := make(map[string]string)

	// Read headers
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break // End of headers
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	// Parse content length
	lengthStr, ok := headers["Content-Length"]
	if !ok {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Content-Length: %w", err)
	}

	// Read body
	body := make([]byte, length)
	_, err = io.ReadFull(s.reader, body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// sendMessage sends a DAP message
func (s *DAPServer) sendMessage(msg any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Set sequence number
	switch m := msg.(type) {
	case *Response:
		m.Seq = int(atomic.AddInt32(&s.seq, 1))
	case *Event:
		m.Seq = int(atomic.AddInt32(&s.seq, 1))
	}

	// Marshal to JSON
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Write headers and body
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	if _, err := s.writer.Write([]byte(header)); err != nil {
		return err
	}
	if _, err := s.writer.Write(body); err != nil {
		return err
	}

	// Flush if supported
	if flusher, ok := s.writer.(interface{ Flush() error }); ok {
		return flusher.Flush()
	}

	return nil
}

// handleMessage processes a single DAP message
func (s *DAPServer) handleMessage(raw []byte) error {
	req, err := ParseRequest(raw)
	if err != nil {
		return err
	}

	switch req.Command {
	case "initialize":
		return s.handleInitialize(req, raw)
	case "launch":
		return s.handleLaunch(req, raw)
	case "attach":
		return s.handleAttach(req, raw)
	case "setBreakpoints":
		return s.handleSetBreakpoints(req, raw)
	case "configurationDone":
		return s.handleConfigurationDone(req)
	case "threads":
		return s.handleThreads(req)
	case "stackTrace":
		return s.handleStackTrace(req, raw)
	case "scopes":
		return s.handleScopes(req, raw)
	case "variables":
		return s.handleVariables(req, raw)
	case "continue":
		return s.handleContinue(req, raw)
	case "next":
		return s.handleNext(req)
	case "stepIn":
		return s.handleStepIn(req)
	case "stepOut":
		return s.handleStepOut(req)
	case "pause":
		return s.handlePause(req)
	case "evaluate":
		return s.handleEvaluate(req, raw)
	case "disconnect":
		return s.handleDisconnect(req)
	default:
		return fmt.Errorf("unknown command: %s", req.Command)
	}
}

// sendErrorResponse sends an error response
func (s *DAPServer) sendErrorResponse(req []byte, message string) {
	var baseReq Request
	json.Unmarshal(req, &baseReq)

	resp := NewResponse(&baseReq, false)
	resp.Message = message
	s.sendMessage(resp)
}

// convertClientToServerLine converts line numbers from client to server format
func (s *DAPServer) convertClientToServerLine(line int) int {
	if s.clientLinesStartAt1 {
		return line
	}
	return line + 1
}

// convertServerToClientLine converts line numbers from server to client format
func (s *DAPServer) convertServerToClientLine(line int) int {
	if s.clientLinesStartAt1 {
		return line
	}
	return line - 1
}

// convertClientToServerColumn converts column numbers from client to server format
func (s *DAPServer) convertClientToServerColumn(column int) int {
	if s.clientColumnsStartAt1 {
		return column
	}
	return column + 1
}

// convertServerToClientColumn converts column numbers from server to client format
func (s *DAPServer) convertServerToClientColumn(column int) int {
	if s.clientColumnsStartAt1 {
		return column
	}
	return column - 1
}

// SendTerminatedEvent sends a terminated event to the client
func (s *DAPServer) SendTerminatedEvent() error {
	event := &TerminatedEvent{
		Event: *NewEvent("terminated"),
	}
	return s.sendMessage(event)
}

// SetProgramFiles sets the files and main expression for attach mode
func (s *DAPServer) SetProgramFiles(files []*FileNode, mainExpr string) {
	s.programFiles = files
	s.mainExpr = mainExpr
}
