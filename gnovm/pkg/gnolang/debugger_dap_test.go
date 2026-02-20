package gnolang

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestDAPMessageParsing(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantErr bool
	}{
		{
			name:    "valid initialize request",
			message: `{"seq":1,"type":"request","command":"initialize","arguments":{"clientID":"test","adapterID":"gno","linesStartAt1":true,"columnsStartAt1":true}}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			message: `{"seq":1,"type":"request",invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseRequest([]byte(tt.message))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && req == nil {
				t.Error("ParseRequest() returned nil request without error")
			}
		})
	}
}

func TestDAPServerMessageFormat(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Create a minimal debugger and machine for testing
	debugger := &Debugger{
		enabled: true,
	}
	machine := &Machine{}

	server := NewDAPServer(debugger, machine)
	server.writer = &buf

	// Test sending a response
	resp := &Response{
		ProtocolMessage: ProtocolMessage{
			Seq:  1,
			Type: "response",
		},
		RequestSeq: 1,
		Success:    true,
		Command:    "initialize",
		Body: Capabilities{
			SupportsConfigurationDoneRequest: true,
		},
	}

	err := server.sendMessage(resp)
	if err != nil {
		t.Fatalf("sendMessage() error = %v", err)
	}

	// Check output format
	output := buf.String()
	if !strings.HasPrefix(output, "Content-Length: ") {
		t.Errorf("Expected output to start with Content-Length header, got: %s", output)
	}

	// Extract JSON body
	parts := strings.Split(output, "\r\n\r\n")
	if len(parts) != 2 {
		t.Fatalf("Expected header and body separated by \\r\\n\\r\\n, got %d parts", len(parts))
	}

	// Verify JSON is valid
	var decoded Response
	if err := json.Unmarshal([]byte(parts[1]), &decoded); err != nil {
		t.Errorf("Failed to decode JSON body: %v", err)
	}
}

func TestDAPLineNumberConversion(t *testing.T) {
	debugger := &Debugger{}
	machine := &Machine{}
	server := NewDAPServer(debugger, machine)

	tests := []struct {
		name            string
		clientStartsAt1 bool
		clientLine      int
		expectedServer  int
		expectedClient  int
	}{
		{
			name:            "client starts at 1",
			clientStartsAt1: true,
			clientLine:      10,
			expectedServer:  10,
			expectedClient:  10,
		},
		{
			name:            "client starts at 0",
			clientStartsAt1: false,
			clientLine:      9,
			expectedServer:  10,
			expectedClient:  9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server.clientLinesStartAt1 = tt.clientStartsAt1

			serverLine := server.convertClientToServerLine(tt.clientLine)
			if serverLine != tt.expectedServer {
				t.Errorf("convertClientToServerLine() = %v, want %v", serverLine, tt.expectedServer)
			}

			clientLine := server.convertServerToClientLine(tt.expectedServer)
			if clientLine != tt.expectedClient {
				t.Errorf("convertServerToClientLine() = %v, want %v", clientLine, tt.expectedClient)
			}
		})
	}
}

type MockWriter struct {
	bytes.Buffer
}

func (m *MockWriter) Flush() error {
	return nil
}

func TestDAPInitializeSequence(t *testing.T) {
	// This is a basic test to ensure the DAP types and handlers compile correctly
	// More comprehensive integration tests would require a full debugging session

	req := &Request{
		ProtocolMessage: ProtocolMessage{
			Seq:  1,
			Type: "request",
		},
		Command: "initialize",
		Arguments: json.RawMessage(`{
			"clientID": "test-client",
			"adapterID": "gno",
			"linesStartAt1": true,
			"columnsStartAt1": true
		}`),
	}

	// Verify we can create proper response
	resp := NewResponse(req, true)
	if resp.RequestSeq != req.Seq {
		t.Errorf("Response RequestSeq = %v, want %v", resp.RequestSeq, req.Seq)
	}
	if resp.Command != req.Command {
		t.Errorf("Response Command = %v, want %v", resp.Command, req.Command)
	}
	if !resp.Success {
		t.Error("Response Success should be true")
	}
}

func ExampleDebugger_ServeDAP() {
	// Create a new machine with debugging enabled
	store := NewStore(nil, nil, nil)
	machine := NewMachineWithOptions(MachineOptions{
		PkgPath: "example",
		Store:   store,
		Debug:   true,
	})

	// Start DAP server
	addr := "localhost:0"
	go func() {
		if err := machine.Debugger.ServeDAP(machine, addr, false, nil, ""); err != nil {
			fmt.Println("DAP server error:", err)
		}
	}()

	// In a real scenario, an IDE would connect to this address
	// and send DAP commands to control debugging
}

func TestDAPDebugLoopIntegration(t *testing.T) {
	// Set up machine with debugging
	store := NewStore(nil, nil, nil)
	machine := NewMachineWithOptions(MachineOptions{
		PkgPath: "test",
		Store:   store,
		Debug:   true,
	})

	// Ensure debugger is properly cleaned up after test
	defer func() {
		machine.Debugger.enabled = false
		machine.Debugger.state = DebugAtExit
		machine.Debugger.dapMode = false
		machine.Debugger.dapServer = nil
	}()

	// Set up DAP server
	machine.Debugger.enabled = true
	machine.Debugger.dapMode = true
	machine.Debugger.state = DebugAtInit

	dapServer := NewDAPServer(&machine.Debugger, machine)
	machine.Debugger.dapServer = dapServer

	// Mock writer for DAP responses
	mockWriter := &MockWriter{}
	dapServer.writer = mockWriter

	tests := []struct {
		name            string
		initialState    DebugState
		dapCommand      string
		expectedState   DebugState
		expectedLastCmd string
	}{
		{
			name:            "DAP continue command",
			initialState:    DebugAtCmd,
			dapCommand:      "continue",
			expectedState:   DebugAtRun,
			expectedLastCmd: "continue",
		},
		{
			name:            "DAP step command",
			initialState:    DebugAtCmd,
			dapCommand:      "step",
			expectedState:   DebugAtRun,
			expectedLastCmd: "step",
		},
		{
			name:            "DAP next command",
			initialState:    DebugAtCmd,
			dapCommand:      "next",
			expectedState:   DebugAtRun,
			expectedLastCmd: "next",
		},
		{
			name:            "DAP stepOut command",
			initialState:    DebugAtCmd,
			dapCommand:      "stepOut",
			expectedState:   DebugAtRun,
			expectedLastCmd: "stepout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			machine.Debugger.state = tt.initialState
			machine.Debugger.lastCmd = ""
			mockWriter.Reset()

			// Create and handle the request
			req := &Request{
				ProtocolMessage: ProtocolMessage{
					Seq:  1,
					Type: "request",
				},
				Command:   tt.dapCommand,
				Arguments: json.RawMessage(`{"threadId":1}`),
			}

			// Handle the command
			var err error
			switch tt.dapCommand {
			case "continue":
				err = dapServer.handleContinue(req, nil)
			case "step":
				req.Command = "stepIn"
				err = dapServer.handleStepIn(req)
			case "next":
				err = dapServer.handleNext(req)
			case "stepOut":
				err = dapServer.handleStepOut(req)
			}

			if err != nil {
				t.Fatalf("Handler failed: %v", err)
			}

			// Check state changes
			if machine.Debugger.state != tt.expectedState {
				t.Errorf("Expected state %v, got %v", tt.expectedState, machine.Debugger.state)
			}
			if machine.Debugger.lastCmd != tt.expectedLastCmd {
				t.Errorf("Expected lastCmd %q, got %q", tt.expectedLastCmd, machine.Debugger.lastCmd)
			}
		})
	}
}

// TestDAPDebugLoopBlocking tests that Debug() loop properly blocks in DAP mode
func TestDAPDebugLoopBlocking(t *testing.T) {
	// Set up machine with debugging
	store := NewStore(nil, nil, nil)
	machine := NewMachineWithOptions(MachineOptions{
		PkgPath: "test",
		Store:   store,
		Debug:   true,
	})

	// Ensure debugger is properly cleaned up after test
	defer func() {
		machine.Debugger.enabled = false
		machine.Debugger.state = DebugAtExit
		machine.Debugger.dapMode = false
		machine.Debugger.dapServer = nil
	}()

	// Create debugger in DAP mode
	machine.Debugger.enabled = true
	machine.Debugger.dapMode = true
	machine.Debugger.state = DebugAtCmd
	machine.Debugger.out = &bytes.Buffer{}

	// Create empty block for testing and initialize machine state properly
	blockStmt := &BlockStmt{Body: []Stmt{}}
	machine.Blocks = []*Block{{Source: blockStmt}}
	// Add a dummy operation to prevent index out of range
	machine.Ops = []Op{OpNoop}
	machine.NumOps = 1

	dapServer := NewDAPServer(&machine.Debugger, machine)
	machine.Debugger.dapServer = dapServer

	// Mock writer
	mockWriter := &MockWriter{}
	dapServer.writer = mockWriter

	// Run Debug() in a goroutine
	debugDone := make(chan bool)
	go func() {
		// This should block when in DAP mode at DebugAtCmd
		// The Debug() function should send a stopped event and transition to DebugAtRun
		machine.Debug()
		debugDone <- true
	}()

	// Give Debug() time to process and send the stopped event
	time.Sleep(100 * time.Millisecond)

	// Check that a stopped event was sent
	output := mockWriter.String()
	if !strings.Contains(output, "stopped") {
		t.Error("Expected stopped event to be sent")
	}

	// Verify state remains at DebugAtCmd (waiting for DAP commands)
	if machine.Debugger.state != DebugAtCmd {
		t.Errorf("Expected state DebugAtCmd, got %v", machine.Debugger.state)
	}

	// Send a DAP continue command to unblock the Debug() loop
	req := &Request{
		ProtocolMessage: ProtocolMessage{Seq: 1, Type: "request"},
		Command:         "continue",
		Arguments:       json.RawMessage(`{"threadId":1}`),
	}

	err := dapServer.handleContinue(req, nil)
	if err != nil {
		t.Fatalf("handleContinue failed: %v", err)
	}

	// Now disable debugger to make Debug() exit
	machine.Debugger.enabled = false

	// Wait for Debug() to complete
	<-debugDone
}

func TestDAPDebugLoopIntegrationWithBreakpoints(t *testing.T) {
	// Set up machine with debugging
	store := NewStore(nil, nil, nil)
	machine := NewMachineWithOptions(MachineOptions{
		PkgPath: "test",
		Store:   store,
		Debug:   true,
	})

	// Ensure debugger is properly cleaned up after test
	defer func() {
		machine.Debugger.enabled = false
		machine.Debugger.state = DebugAtExit
		machine.Debugger.dapMode = false
		machine.Debugger.dapServer = nil
	}()

	// Set up DAP server
	machine.Debugger.enabled = true
	machine.Debugger.dapMode = true
	machine.Debugger.state = DebugAtCmd

	// Set a breakpoint
	machine.Debugger.breakpoints = []Location{
		{File: "test.gno", Span: Span{Pos: Pos{Line: 10}}},
	}
	machine.Debugger.loc = Location{File: "test.gno", Span: Span{Pos: Pos{Line: 10}}}
	// Set prevLoc to something different so atBreak() returns true
	machine.Debugger.prevLoc = Location{File: "test.gno", Span: Span{Pos: Pos{Line: 9}}}

	// Initialize machine state
	blockStmt := &BlockStmt{Body: []Stmt{}}
	machine.Blocks = []*Block{{Source: blockStmt}}
	machine.Ops = []Op{OpNoop}
	machine.NumOps = 1

	dapServer := NewDAPServer(&machine.Debugger, machine)
	machine.Debugger.dapServer = dapServer

	// Mock writer
	mockWriter := &MockWriter{}
	dapServer.writer = mockWriter

	// Test that stopped event is sent with breakpoint reason
	debugDone := make(chan bool)
	go func() {
		machine.Debug()
		debugDone <- true
	}()

	// Wait for stopped event
	time.Sleep(50 * time.Millisecond)

	output := mockWriter.String()
	if !strings.Contains(output, "stopped") {
		t.Error("Expected stopped event to be sent")
	}
	if !strings.Contains(output, "breakpoint") {
		t.Logf("Output: %q", output)
		t.Error("Expected stopped event with breakpoint reason")
	}

	// Verify state remains at DebugAtCmd waiting for DAP command
	if machine.Debugger.state != DebugAtCmd {
		t.Errorf("Expected state DebugAtCmd, got %v", machine.Debugger.state)
	}

	// Send continue command
	req := &Request{
		ProtocolMessage: ProtocolMessage{Seq: 1, Type: "request"},
		Command:         "continue",
		Arguments:       json.RawMessage(`{"threadId":1}`),
	}

	mockWriter.Reset()
	err := dapServer.handleContinue(req, nil)
	if err != nil {
		t.Fatalf("handleContinue failed: %v", err)
	}

	// Verify state changed to DebugAtRun
	if machine.Debugger.state != DebugAtRun {
		t.Errorf("Expected state DebugAtRun after continue, got %v", machine.Debugger.state)
	}

	// Terminate debugging
	machine.Debugger.enabled = false
	<-debugDone
}

func TestDAPStoppedEventReasons(t *testing.T) {
	debugger := &Debugger{}
	machine := &Machine{}
	server := NewDAPServer(debugger, machine)

	mockWriter := &MockWriter{}
	server.writer = mockWriter

	tests := []struct {
		reason      string
		description string
	}{
		{"breakpoint", "Hit breakpoint at line 10"},
		{"step", "Stepped to next line"},
		{"pause", "Paused by user request"},
		{"exception", "Runtime error occurred"},
		{"entry", "Program started"},
	}

	for _, tt := range tests {
		t.Run(tt.reason, func(t *testing.T) {
			mockWriter.Reset()

			err := server.SendStoppedEvent(tt.reason, tt.description)
			if err != nil {
				t.Fatalf("SendStoppedEvent failed: %v", err)
			}

			// Parse the output
			output := mockWriter.String()
			parts := strings.Split(output, "\r\n\r\n")
			if len(parts) != 2 {
				t.Fatal("Invalid output format")
			}

			var event StoppedEvent
			if err := json.Unmarshal([]byte(parts[1]), &event); err != nil {
				t.Fatalf("Failed to parse event: %v", err)
			}

			if event.Body.Reason != tt.reason {
				t.Errorf("Expected reason %q, got %q", tt.reason, event.Body.Reason)
			}
			if event.Body.Description != tt.description {
				t.Errorf("Expected description %q, got %q", tt.description, event.Body.Description)
			}
		})
	}
}
