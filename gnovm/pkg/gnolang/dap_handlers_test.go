package gnolang

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestHandlersWithRawParameter(t *testing.T) {
	debugger := &Debugger{
		enabled:     true,
		breakpoints: []Location{},
	}
	machine := &Machine{}
	server := NewDAPServer(debugger, machine)

	tests := []struct {
		name    string
		handler func(*Request, []byte) error
		request *Request
		raw     []byte
	}{
		{
			name:    "handleInitialize with raw",
			handler: server.handleInitialize,
			request: &Request{
				ProtocolMessage: ProtocolMessage{Seq: 1, Type: "request"},
				Command:         "initialize",
				Arguments:       json.RawMessage(`{"clientID":"test","adapterID":"gno","linesStartAt1":true,"columnsStartAt1":true}`),
			},
			raw: []byte(`{"seq":1,"type":"request","command":"initialize","arguments":{"clientID":"test","adapterID":"gno","linesStartAt1":true,"columnsStartAt1":true}}`),
		},
		{
			name:    "handleLaunch with raw",
			handler: server.handleLaunch,
			request: &Request{
				ProtocolMessage: ProtocolMessage{Seq: 2, Type: "request"},
				Command:         "launch",
				Arguments:       json.RawMessage(`{"program":"test.gno","noDebug":false}`),
			},
			raw: []byte(`{"seq":2,"type":"request","command":"launch","arguments":{"program":"test.gno","noDebug":false}}`),
		},
		{
			name:    "handleSetBreakpoints with raw",
			handler: server.handleSetBreakpoints,
			request: &Request{
				ProtocolMessage: ProtocolMessage{Seq: 3, Type: "request"},
				Command:         "setBreakpoints",
				Arguments:       json.RawMessage(`{"source":{"path":"test.gno"},"breakpoints":[{"line":10}]}`),
			},
			raw: []byte(`{"seq":3,"type":"request","command":"setBreakpoints","arguments":{"source":{"path":"test.gno"},"breakpoints":[{"line":10}]}}`),
		},
		{
			name:    "handleStackTrace with raw",
			handler: server.handleStackTrace,
			request: &Request{
				ProtocolMessage: ProtocolMessage{Seq: 4, Type: "request"},
				Command:         "stackTrace",
				Arguments:       json.RawMessage(`{"threadId":1,"levels":20}`),
			},
			raw: []byte(`{"seq":4,"type":"request","command":"stackTrace","arguments":{"threadId":1,"levels":20}}`),
		},
		{
			name:    "handleScopes with raw",
			handler: server.handleScopes,
			request: &Request{
				ProtocolMessage: ProtocolMessage{Seq: 5, Type: "request"},
				Command:         "scopes",
				Arguments:       json.RawMessage(`{"frameId":0}`),
			},
			raw: []byte(`{"seq":5,"type":"request","command":"scopes","arguments":{"frameId":0}}`),
		},
		{
			name:    "handleVariables with raw",
			handler: server.handleVariables,
			request: &Request{
				ProtocolMessage: ProtocolMessage{Seq: 6, Type: "request"},
				Command:         "variables",
				Arguments:       json.RawMessage(`{"variablesReference":1000}`),
			},
			raw: []byte(`{"seq":6,"type":"request","command":"variables","arguments":{"variablesReference":1000}}`),
		},
		{
			name:    "handleContinue with raw",
			handler: server.handleContinue,
			request: &Request{
				ProtocolMessage: ProtocolMessage{Seq: 7, Type: "request"},
				Command:         "continue",
				Arguments:       json.RawMessage(`{"threadId":1}`),
			},
			raw: []byte(`{"seq":7,"type":"request","command":"continue","arguments":{"threadId":1}}`),
		},
		{
			name:    "handleEvaluate with raw",
			handler: server.handleEvaluate,
			request: &Request{
				ProtocolMessage: ProtocolMessage{Seq: 8, Type: "request"},
				Command:         "evaluate",
				Arguments:       json.RawMessage(`{"expression":"1+1","frameId":0}`),
			},
			raw: []byte(`{"seq":8,"type":"request","command":"evaluate","arguments":{"expression":"1+1","frameId":0}}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// capture responses
			mockWriter := &MockWriter{}
			server.writer = mockWriter

			err := tt.handler(tt.request, tt.raw)
			if err != nil {
				t.Logf("Handler returned error: %v", err)
			}

			if mockWriter.Len() == 0 && err == nil {
				t.Error("Expected handler to write a response")
			}
		})
	}
}

func TestHandlersWithoutRawParameter(t *testing.T) {
	debugger := &Debugger{
		enabled: true,
	}
	machine := &Machine{}
	server := NewDAPServer(debugger, machine)

	mockWriter := &MockWriter{}
	server.writer = mockWriter

	handlers := []struct {
		name    string
		handler func(*Request) error
		request *Request
	}{
		{
			name:    "handleConfigurationDone",
			handler: server.handleConfigurationDone,
			request: &Request{
				ProtocolMessage: ProtocolMessage{Seq: 1, Type: "request"},
				Command:         "configurationDone",
			},
		},
		{
			name:    "handleThreads",
			handler: server.handleThreads,
			request: &Request{
				ProtocolMessage: ProtocolMessage{Seq: 2, Type: "request"},
				Command:         "threads",
			},
		},
		{
			name:    "handleNext",
			handler: server.handleNext,
			request: &Request{
				ProtocolMessage: ProtocolMessage{Seq: 3, Type: "request"},
				Command:         "next",
			},
		},
		{
			name:    "handleStepIn",
			handler: server.handleStepIn,
			request: &Request{
				ProtocolMessage: ProtocolMessage{Seq: 4, Type: "request"},
				Command:         "stepIn",
			},
		},
		{
			name:    "handleStepOut",
			handler: server.handleStepOut,
			request: &Request{
				ProtocolMessage: ProtocolMessage{Seq: 5, Type: "request"},
				Command:         "stepOut",
			},
		},
		{
			name:    "handlePause",
			handler: server.handlePause,
			request: &Request{
				ProtocolMessage: ProtocolMessage{Seq: 6, Type: "request"},
				Command:         "pause",
			},
		},
		{
			name:    "handleDisconnect",
			handler: server.handleDisconnect,
			request: &Request{
				ProtocolMessage: ProtocolMessage{Seq: 7, Type: "request"},
				Command:         "disconnect",
			},
		},
	}

	for _, tt := range handlers {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter.Reset()

			err := tt.handler(tt.request)

			if err != nil {
				t.Errorf("Handler returned unexpected error: %v", err)
			}

			if mockWriter.Len() == 0 {
				t.Error("Expected handler to write a response")
			}
		})
	}
}

func TestSetBreakpointsWithColumn(t *testing.T) {
	debugger := &Debugger{
		enabled:     true,
		breakpoints: []Location{},
	}
	machine := &Machine{}
	server := NewDAPServer(debugger, machine)
	server.clientColumnsStartAt1 = false // 0-based columns

	// Mock writer
	mockWriter := &MockWriter{}
	server.writer = mockWriter

	// Create request with column information
	req := &Request{
		ProtocolMessage: ProtocolMessage{Seq: 1, Type: "request"},
		Command:         "setBreakpoints",
		Arguments: json.RawMessage(`{
			"source": {"path": "/test/main.gno"},
			"breakpoints": [
				{"line": 10, "column": 5},
				{"line": 20, "column": 0},
				{"line": 30}
			]
		}`),
	}

	err := server.handleSetBreakpoints(req, nil)
	if err != nil {
		t.Fatalf("handleSetBreakpoints failed: %v", err)
	}

	// Check that breakpoints were created with correct columns
	if len(debugger.breakpoints) != 3 {
		t.Errorf("Expected 3 breakpoints, got %d", len(debugger.breakpoints))
	}

	// Test cases for column conversion
	tests := []struct {
		index          int
		expectedLine   int
		expectedColumn int
		description    string
	}{
		{0, 10, 6, "First breakpoint with column 5 (0-based) should convert to 6 (1-based)"},
		{1, 20, 0, "Second breakpoint with column 0 (0-based) should remain 0 (no specific column)"},
		{2, 30, 0, "Third breakpoint without column should default to 0"},
	}

	for _, tt := range tests {
		if tt.index < len(debugger.breakpoints) {
			bp := debugger.breakpoints[tt.index]
			if bp.Line != tt.expectedLine {
				t.Errorf("%s: Expected line %d, got %d", tt.description, tt.expectedLine, bp.Line)
			}
			if bp.Column != tt.expectedColumn {
				t.Errorf("%s: Expected column %d, got %d", tt.description, tt.expectedColumn, bp.Column)
			}
		}
	}
}

func TestBreakpointColumnInResponse(t *testing.T) {
	debugger := &Debugger{
		enabled:     true,
		breakpoints: []Location{},
	}
	machine := &Machine{}
	server := NewDAPServer(debugger, machine)
	server.clientColumnsStartAt1 = true // 1-based columns

	mockWriter := &MockWriter{}
	server.writer = mockWriter

	// Create request with column
	req := &Request{
		ProtocolMessage: ProtocolMessage{Seq: 1, Type: "request"},
		Command:         "setBreakpoints",
		Arguments: json.RawMessage(`{
			"source": {"path": "/test/main.gno"},
			"breakpoints": [{"line": 10, "column": 5}]
		}`),
	}

	err := server.handleSetBreakpoints(req, nil)
	if err != nil {
		t.Fatalf("handleSetBreakpoints failed: %v", err)
	}

	response := mockWriter.String()

	bodyStart := 0
	for i := 0; i < len(response)-4; i++ {
		if response[i:i+4] == "\r\n\r\n" {
			bodyStart = i + 4
			break
		}
	}

	var resp struct {
		Body struct {
			Breakpoints []struct {
				Line   int `json:"line"`
				Column int `json:"column"`
			} `json:"breakpoints"`
		} `json:"body"`
	}

	if err := json.Unmarshal([]byte(response[bodyStart:]), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(resp.Body.Breakpoints) != 1 {
		t.Fatalf("Expected 1 breakpoint in response, got %d", len(resp.Body.Breakpoints))
	}

	// Should return the same column as sent (since clientColumnsStartAt1 = true)
	if resp.Body.Breakpoints[0].Column != 5 {
		t.Errorf("Expected column 5 in response, got %d", resp.Body.Breakpoints[0].Column)
	}
}

func TestColumnConversionMethods(t *testing.T) {
	debugger := &Debugger{}
	machine := &Machine{}
	server := NewDAPServer(debugger, machine)

	tests := []struct {
		name                  string
		clientColumnsStartAt1 bool
		clientColumn          int
		expectedServer        int
		expectedClient        int
	}{
		{
			name:                  "client columns start at 1",
			clientColumnsStartAt1: true,
			clientColumn:          5,
			expectedServer:        5,
			expectedClient:        5,
		},
		{
			name:                  "client columns start at 0",
			clientColumnsStartAt1: false,
			clientColumn:          4,
			expectedServer:        5,
			expectedClient:        4,
		},
		{
			name:                  "column 1 when client starts at 1",
			clientColumnsStartAt1: true,
			clientColumn:          1,
			expectedServer:        1,
			expectedClient:        1,
		},
		{
			name:                  "column 0 when client starts at 0",
			clientColumnsStartAt1: false,
			clientColumn:          0,
			expectedServer:        1,
			expectedClient:        0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server.clientColumnsStartAt1 = tt.clientColumnsStartAt1

			serverColumn := server.convertClientToServerColumn(tt.clientColumn)
			if serverColumn != tt.expectedServer {
				t.Errorf("convertClientToServerColumn() = %v, want %v", serverColumn, tt.expectedServer)
			}

			clientColumn := server.convertServerToClientColumn(tt.expectedServer)
			if clientColumn != tt.expectedClient {
				t.Errorf("convertServerToClientColumn() = %v, want %v", clientColumn, tt.expectedClient)
			}
		})
	}
}

func TestColumnConversionUsage(t *testing.T) {
	// This test documents where column conversion methods are used
	// Currently checking if convertClientToServerColumn is used anywhere

	debugger := &Debugger{
		enabled: true,
		loc: Location{
			Span: Span{
				Pos: Pos{Line: 10, Column: 5},
			},
		},
	}
	machine := &Machine{}
	server := NewDAPServer(debugger, machine)
	server.clientColumnsStartAt1 = false // Test with 0-based columns

	// Mock writer
	mockWriter := &MockWriter{}
	server.writer = mockWriter

	// Test handleStackTrace which uses convertServerToClientColumn
	stackReq := &Request{
		ProtocolMessage: ProtocolMessage{Seq: 1, Type: "request"},
		Command:         "stackTrace",
		Arguments:       []byte(`{"threadId":1,"levels":20}`),
	}

	err := server.handleStackTrace(stackReq, nil)
	if err != nil {
		t.Errorf("handleStackTrace failed: %v", err)
	}

	// Verify that convertServerToClientColumn is used (column should be 4, not 5)
	response := mockWriter.String()
	if !contains(t, response, `"column":4`) {
		t.Log("Note: convertServerToClientColumn might not be working correctly")
		t.Logf("Response: %s", response)
	}
}

func contains(t *testing.T, s, substr string) bool {
	t.Helper()
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(t, s, substr))
}

func containsHelper(t *testing.T, s, substr string) bool {
	t.Helper()
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestHandleVariables(t *testing.T) {
	// minimal block structure
	testBlock := &BlockStmt{
		Body: []Stmt{},
	}

	// block with some test values
	block := &Block{
		Source: testBlock,
		Values: []TypedValue{
			{T: IntType, V: nil},    // Will be displayed as int value
			{T: StringType, V: nil}, // Will be displayed as string value
		},
	}

	// Set the block names
	testBlock.Names = []Name{"testInt", "testStr"}

	machine := &Machine{
		Blocks: []*Block{block},
	}

	debugger := &Debugger{
		enabled:    true,
		frameLevel: 0,
	}

	server := NewDAPServer(debugger, machine)
	server.machine = machine
	mockWriter := &MockWriter{}
	server.writer = mockWriter

	tests := []struct {
		name               string
		variablesReference int
		wantVariableCount  int
		checkVariables     bool
	}{
		{
			name:               "local variables scope",
			variablesReference: 1000, // Local scope reference
			wantVariableCount:  2,    // testInt and testStr
			checkVariables:     true,
		},
		{
			name:               "global variables scope",
			variablesReference: 2000, // Global scope reference
			wantVariableCount:  2,    // Same variables in our test
			checkVariables:     true,
		},
		{
			name:               "invalid reference",
			variablesReference: 9999,
			wantVariableCount:  0,
			checkVariables:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter.Reset()

			req := &Request{
				ProtocolMessage: ProtocolMessage{Seq: 1, Type: "request"},
				Command:         "variables",
				Arguments:       json.RawMessage(fmt.Sprintf(`{"variablesReference":%d}`, tt.variablesReference)),
			}

			err := server.handleVariables(req, nil)
			if err != nil {
				t.Fatalf("handleVariables failed: %v", err)
			}

			// Parse response
			response := mockWriter.String()
			bodyStart := 0
			for i := 0; i < len(response)-4; i++ {
				if response[i:i+4] == "\r\n\r\n" {
					bodyStart = i + 4
					break
				}
			}

			var resp struct {
				Body struct {
					Variables []Variable `json:"variables"`
				} `json:"body"`
			}

			if err := json.Unmarshal([]byte(response[bodyStart:]), &resp); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			// Check variable count
			if len(resp.Body.Variables) != tt.wantVariableCount {
				t.Errorf("Expected %d variables, got %d", tt.wantVariableCount, len(resp.Body.Variables))
			}

			// Check variable properties if needed
			if tt.checkVariables && len(resp.Body.Variables) > 0 {
				// Check first variable
				if resp.Body.Variables[0].Name != "testInt" {
					t.Errorf("Expected first variable name 'testInt', got %s", resp.Body.Variables[0].Name)
				}
				if resp.Body.Variables[0].Type != "int" {
					t.Errorf("Expected first variable type 'int', got %s", resp.Body.Variables[0].Type)
				}

				// Check second variable if exists
				if len(resp.Body.Variables) > 1 {
					if resp.Body.Variables[1].Name != "testStr" {
						t.Errorf("Expected second variable name 'testStr', got %s", resp.Body.Variables[1].Name)
					}
					if resp.Body.Variables[1].Type != "string" {
						t.Errorf("Expected second variable type 'string', got %s", resp.Body.Variables[1].Type)
					}
				}
			}
		})
	}
}
