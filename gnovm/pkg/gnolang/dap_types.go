package gnolang

import (
	"encoding/json"
	"fmt"
)

// DAP Base Protocol Types

// ProtocolMessage is the base class of requests, responses, and events
type ProtocolMessage struct {
	Seq  int    `json:"seq"`
	Type string `json:"type"` // "request", "response", "event"
}

// Request represents a client or debug adapter initiated request
type Request struct {
	ProtocolMessage
	Command   string          `json:"command"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// Response represents a response to a request
type Response struct {
	ProtocolMessage
	RequestSeq int            `json:"request_seq"`
	Success    bool           `json:"success"`
	Command    string         `json:"command"`
	Message    string         `json:"message,omitempty"`
	Body       any            `json:"body,omitempty"`
	ErrorBody  *ErrorResponse `json:"error,omitempty"`
}

// Event represents a debug adapter initiated event
type Event struct {
	ProtocolMessage
	Event string `json:"event"`
	Body  any    `json:"body,omitempty"`
}

// ErrorResponse contains error information
type ErrorResponse struct {
	ID        int            `json:"id"`
	Format    string         `json:"format"`
	Variables map[string]any `json:"variables,omitempty"`
	ShowUser  bool           `json:"showUser,omitempty"`
	URL       string         `json:"url,omitempty"`
	URLLabel  string         `json:"urlLabel,omitempty"`
}

// DAP Request Types

// InitializeRequest is the first request sent by the client
type InitializeRequest struct {
	Request
	Arguments InitializeArguments `json:"arguments"`
}

type InitializeArguments struct {
	ClientID                     string `json:"clientID,omitempty"`
	ClientName                   string `json:"clientName,omitempty"`
	AdapterID                    string `json:"adapterID"`
	Locale                       string `json:"locale,omitempty"`
	LinesStartAt1                bool   `json:"linesStartAt1"`
	ColumnsStartAt1              bool   `json:"columnsStartAt1"`
	PathFormat                   string `json:"pathFormat,omitempty"`
	SupportsVariableType         bool   `json:"supportsVariableType,omitempty"`
	SupportsVariablePaging       bool   `json:"supportsVariablePaging,omitempty"`
	SupportsRunInTerminalRequest bool   `json:"supportsRunInTerminalRequest,omitempty"`
	SupportsMemoryReferences     bool   `json:"supportsMemoryReferences,omitempty"`
}

// Capabilities describes the debug adapter's capabilities
type Capabilities struct {
	SupportsConfigurationDoneRequest   bool                         `json:"supportsConfigurationDoneRequest,omitempty"`
	SupportsFunctionBreakpoints        bool                         `json:"supportsFunctionBreakpoints,omitempty"`
	SupportsConditionalBreakpoints     bool                         `json:"supportsConditionalBreakpoints,omitempty"`
	SupportsHitConditionalBreakpoints  bool                         `json:"supportsHitConditionalBreakpoints,omitempty"`
	SupportsEvaluateForHovers          bool                         `json:"supportsEvaluateForHovers,omitempty"`
	ExceptionBreakpointFilters         []ExceptionBreakpointsFilter `json:"exceptionBreakpointFilters,omitempty"`
	SupportsStepBack                   bool                         `json:"supportsStepBack,omitempty"`
	SupportsSetVariable                bool                         `json:"supportsSetVariable,omitempty"`
	SupportsRestartFrame               bool                         `json:"supportsRestartFrame,omitempty"`
	SupportsGotoTargetsRequest         bool                         `json:"supportsGotoTargetsRequest,omitempty"`
	SupportsStepInTargetsRequest       bool                         `json:"supportsStepInTargetsRequest,omitempty"`
	SupportsCompletionsRequest         bool                         `json:"supportsCompletionsRequest,omitempty"`
	SupportsModulesRequest             bool                         `json:"supportsModulesRequest,omitempty"`
	SupportsRestartRequest             bool                         `json:"supportsRestartRequest,omitempty"`
	SupportsExceptionOptions           bool                         `json:"supportsExceptionOptions,omitempty"`
	SupportsValueFormattingOptions     bool                         `json:"supportsValueFormattingOptions,omitempty"`
	SupportsExceptionInfoRequest       bool                         `json:"supportsExceptionInfoRequest,omitempty"`
	SupportTerminateDebuggee           bool                         `json:"supportTerminateDebuggee,omitempty"`
	SupportsDelayedStackTraceLoading   bool                         `json:"supportsDelayedStackTraceLoading,omitempty"`
	SupportsLoadedSourcesRequest       bool                         `json:"supportsLoadedSourcesRequest,omitempty"`
	SupportsLogPoints                  bool                         `json:"supportsLogPoints,omitempty"`
	SupportsTerminateThreadsRequest    bool                         `json:"supportsTerminateThreadsRequest,omitempty"`
	SupportsSetExpression              bool                         `json:"supportsSetExpression,omitempty"`
	SupportsTerminateRequest           bool                         `json:"supportsTerminateRequest,omitempty"`
	SupportsDataBreakpoints            bool                         `json:"supportsDataBreakpoints,omitempty"`
	SupportsReadMemoryRequest          bool                         `json:"supportsReadMemoryRequest,omitempty"`
	SupportsDisassembleRequest         bool                         `json:"supportsDisassembleRequest,omitempty"`
	SupportsCancelRequest              bool                         `json:"supportsCancelRequest,omitempty"`
	SupportsBreakpointLocationsRequest bool                         `json:"supportsBreakpointLocationsRequest,omitempty"`
}

type ExceptionBreakpointsFilter struct {
	Filter  string `json:"filter"`
	Label   string `json:"label"`
	Default bool   `json:"default,omitempty"`
}

// LaunchRequest is sent to start the debuggee
type LaunchRequest struct {
	Request
	Arguments LaunchArguments `json:"arguments"`
}

type LaunchArguments struct {
	NoDebug bool              `json:"noDebug,omitempty"`
	Program string            `json:"program"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Cwd     string            `json:"cwd,omitempty"`
}

// AttachRequest is sent to attach to a running process
type AttachRequest struct {
	Request
	Arguments AttachArguments `json:"arguments"`
}

type AttachArguments struct {
	ProcessID int    `json:"processId,omitempty"`
	Port      int    `json:"port,omitempty"`
	Host      string `json:"host,omitempty"`
}

// SetBreakpointsRequest is sent to set breakpoints
type SetBreakpointsRequest struct {
	Request
	Arguments SetBreakpointsArguments `json:"arguments"`
}

type SetBreakpointsArguments struct {
	Source      Source             `json:"source"`
	Breakpoints []SourceBreakpoint `json:"breakpoints,omitempty"`
}

type Source struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

type SourceBreakpoint struct {
	Line         int    `json:"line"`
	Column       int    `json:"column,omitempty"`
	Condition    string `json:"condition,omitempty"`
	HitCondition string `json:"hitCondition,omitempty"`
	LogMessage   string `json:"logMessage,omitempty"`
}

type Breakpoint struct {
	ID       int    `json:"id,omitempty"`
	Verified bool   `json:"verified"`
	Message  string `json:"message,omitempty"`
	Source   Source `json:"source,omitempty"`
	Line     int    `json:"line,omitempty"`
	Column   int    `json:"column,omitempty"`
}

// ContinueRequest is sent to resume execution
type ContinueRequest struct {
	Request
	Arguments ContinueArguments `json:"arguments"`
}

type ContinueArguments struct {
	ThreadID int `json:"threadId"`
}

// NextRequest is sent to step over
type NextRequest struct {
	Request
	Arguments StepArguments `json:"arguments"`
}

// StepInRequest is sent to step into
type StepInRequest struct {
	Request
	Arguments StepArguments `json:"arguments"`
}

// StepOutRequest is sent to step out
type StepOutRequest struct {
	Request
	Arguments StepArguments `json:"arguments"`
}

type StepArguments struct {
	ThreadID int `json:"threadId"`
}

// StackTraceRequest is sent to get the stack trace
type StackTraceRequest struct {
	Request
	Arguments StackTraceArguments `json:"arguments"`
}

type StackTraceArguments struct {
	ThreadID   int    `json:"threadId"`
	StartFrame int    `json:"startFrame,omitempty"`
	Levels     int    `json:"levels,omitempty"`
	Format     string `json:"format,omitempty"`
}

type StackFrame struct {
	ID                          int    `json:"id"`
	Name                        string `json:"name"`
	Source                      Source `json:"source,omitempty"`
	Line                        int    `json:"line"`
	Column                      int    `json:"column"`
	EndLine                     int    `json:"endLine,omitempty"`
	EndColumn                   int    `json:"endColumn,omitempty"`
	CanRestart                  bool   `json:"canRestart,omitempty"`
	InstructionPointerReference string `json:"instructionPointerReference,omitempty"`
}

// EvaluateRequest is sent to evaluate an expression
type EvaluateRequest struct {
	Request
	Arguments EvaluateArguments `json:"arguments"`
}

type EvaluateArguments struct {
	Expression string `json:"expression"`
	FrameID    int    `json:"frameId,omitempty"`
	Context    string `json:"context,omitempty"`
	Format     string `json:"format,omitempty"`
}

// ScopesRequest is sent to get scopes for a stack frame
type ScopesRequest struct {
	Request
	Arguments ScopesArguments `json:"arguments"`
}

type ScopesArguments struct {
	FrameID int `json:"frameId"`
}

type Scope struct {
	Name               string `json:"name"`
	VariablesReference int    `json:"variablesReference"`
	NamedVariables     int    `json:"namedVariables,omitempty"`
	IndexedVariables   int    `json:"indexedVariables,omitempty"`
	Expensive          bool   `json:"expensive"`
}

// VariablesRequest is sent to get variables in a scope
type VariablesRequest struct {
	Request
	Arguments VariablesArguments `json:"arguments"`
}

type VariablesArguments struct {
	VariablesReference int    `json:"variablesReference"`
	Filter             string `json:"filter,omitempty"`
	Start              int    `json:"start,omitempty"`
	Count              int    `json:"count,omitempty"`
}

type Variable struct {
	Name               string `json:"name"`
	Value              string `json:"value"`
	Type               string `json:"type,omitempty"`
	VariablesReference int    `json:"variablesReference"`
	NamedVariables     int    `json:"namedVariables,omitempty"`
	IndexedVariables   int    `json:"indexedVariables,omitempty"`
}

// DAP Event Types

// StoppedEvent is sent when execution stops
type StoppedEvent struct {
	Event
	Body StoppedEventBody `json:"body"`
}

type StoppedEventBody struct {
	Reason            string `json:"reason"`
	Description       string `json:"description,omitempty"`
	ThreadID          int    `json:"threadId,omitempty"`
	PreserveFocusHint bool   `json:"preserveFocusHint,omitempty"`
	Text              string `json:"text,omitempty"`
	AllThreadsStopped bool   `json:"allThreadsStopped,omitempty"`
}

// ContinuedEvent is sent when execution continues
type ContinuedEvent struct {
	Event
	Body ContinuedEventBody `json:"body"`
}

type ContinuedEventBody struct {
	ThreadID            int  `json:"threadId"`
	AllThreadsContinued bool `json:"allThreadsContinued,omitempty"`
}

// TerminatedEvent is sent when debugging terminates
type TerminatedEvent struct {
	Event
	Body TerminatedEventBody `json:"body,omitempty"`
}

type TerminatedEventBody struct {
	Restart any `json:"restart,omitempty"`
}

// InitializedEvent is sent when the debug adapter is ready
type InitializedEvent struct {
	Event
}

// OutputEvent is sent to output console messages
type OutputEvent struct {
	Event
	Body OutputEventBody `json:"body"`
}

type OutputEventBody struct {
	Category string `json:"category,omitempty"`
	Output   string `json:"output"`
	Source   Source `json:"source,omitempty"`
	Line     int    `json:"line,omitempty"`
	Column   int    `json:"column,omitempty"`
}

// Helper functions

// NewResponse creates a new response for a request
func NewResponse(req *Request, success bool) *Response {
	return &Response{
		ProtocolMessage: ProtocolMessage{
			Seq:  0, // Will be set by the server
			Type: "response",
		},
		RequestSeq: req.Seq,
		Success:    success,
		Command:    req.Command,
	}
}

// NewEvent creates a new event
func NewEvent(eventType string) *Event {
	return &Event{
		ProtocolMessage: ProtocolMessage{
			Seq:  0, // Will be set by the server
			Type: "event",
		},
		Event: eventType,
	}
}

// ParseRequest parses a raw message into a specific request type
func ParseRequest(raw []byte) (*Request, error) {
	var req Request
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}
	return &req, nil
}
