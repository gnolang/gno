package gnolang

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
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
		if err := machine.Debugger.ServeDAP(machine, addr); err != nil {
			fmt.Println("DAP server error:", err)
		}
	}()

	// In a real scenario, an IDE would connect to this address
	// and send DAP commands to control debugging
}
