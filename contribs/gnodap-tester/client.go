package client

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"path/filepath"
	"testing"

	"github.com/google/go-dap"
)

type DAPClient struct {
	conn   net.Conn
	reader *bufio.Reader
	seq    int
}

func NewDAPClient(addr string) *DAPClient {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to connect to %s: %v", addr, err)
	}
	return &DAPClient{
		conn:   conn,
		reader: bufio.NewReader(conn),
		seq:    1,
	}
}

// Close closes the client connection.
func (c *DAPClient) Close() {
	c.conn.Close()
}

func (c *DAPClient) send(request dap.Message) {
	dap.WriteProtocolMessage(c.conn, request)
}

func (c *DAPClient) ReadMessage() (dap.Message, error) {
	return dap.ReadProtocolMessage(c.reader)
}

func (c *DAPClient) ExpectMessage(t *testing.T) dap.Message {
	t.Helper()
	m, err := dap.ReadProtocolMessage(c.reader)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func (c *DAPClient) InitializeRequest() {
	request := &dap.InitializeRequest{Request: *c.newRequest("initialize")}
	request.Arguments = dap.InitializeRequestArguments{
		AdapterID:  "gno",
		PathFormat: "path",
		Locale:     "en-us",
	}
	c.send(request)
}

func (c *DAPClient) InitializeRequestWithArgs(args dap.InitializeRequestArguments) {
	request := &dap.InitializeRequest{Request: *c.newRequest("initialize")}
	request.Arguments = args
	c.send(request)
}

func toRawMessage(in interface{}) json.RawMessage {
	out, _ := json.Marshal(in)
	return out
}

func (c *DAPClient) LaunchRequest(mode, program string, stopOnEntry bool) {
	request := &dap.LaunchRequest{Request: *c.newRequest("launch")}
	request.Arguments = toRawMessage(map[string]interface{}{
		"request":     "launch",
		"mode":        mode,
		"program":     program,
		"stopOnEntry": stopOnEntry,
	})
	c.send(request)
}

func (c *DAPClient) LaunchRequestWithArgs(arguments map[string]interface{}) {
	request := &dap.LaunchRequest{Request: *c.newRequest("launch")}
	request.Arguments = toRawMessage(arguments)
	c.send(request)
}

func (c *DAPClient) AttachRequest(arguments map[string]interface{}) {
	request := &dap.AttachRequest{Request: *c.newRequest("attach")}
	request.Arguments = toRawMessage(arguments)
	c.send(request)
}

func (c *DAPClient) DisconnectRequest() {
	request := &dap.DisconnectRequest{Request: *c.newRequest("disconnect")}
	c.send(request)
}

func (c *DAPClient) DisconnectRequestWithKillOption(kill bool) {
	request := &dap.DisconnectRequest{Request: *c.newRequest("disconnect")}
	request.Arguments.TerminateDebuggee = kill
	c.send(request)
}

func (c *DAPClient) SetBreakpointsRequest(file string, lines []int) {
	c.SetBreakpointsRequestWithArgs(file, lines, nil, nil, nil)
}

func (c *DAPClient) SetBreakpointsRequestWithArgs(file string, lines []int, conditions, hitConditions, logMessages map[int]string) {
	request := &dap.SetBreakpointsRequest{Request: *c.newRequest("setBreakpoints")}
	request.Arguments = dap.SetBreakpointsArguments{
		Source: dap.Source{
			Name: filepath.Base(file),
			Path: file,
		},
		Breakpoints: make([]dap.SourceBreakpoint, len(lines)),
	}
	for i, l := range lines {
		request.Arguments.Breakpoints[i].Line = l
		if cond, ok := conditions[l]; ok {
			request.Arguments.Breakpoints[i].Condition = cond
		}
		if hitCond, ok := hitConditions[l]; ok {
			request.Arguments.Breakpoints[i].HitCondition = hitCond
		}
		if logMessage, ok := logMessages[l]; ok {
			request.Arguments.Breakpoints[i].LogMessage = logMessage
		}
	}
	c.send(request)
}

func (c *DAPClient) SetExceptionBreakpointsRequest() {
	request := &dap.SetBreakpointsRequest{Request: *c.newRequest("setExceptionBreakpoints")}
	c.send(request)
}

func (c *DAPClient) ConfigurationDoneRequest() {
	request := &dap.ConfigurationDoneRequest{Request: *c.newRequest("configurationDone")}
	c.send(request)
}

// ContinueRequest sends a 'continue' request.
func (c *DAPClient) ContinueRequest(thread int) {
	request := &dap.ContinueRequest{Request: *c.newRequest("continue")}
	request.Arguments.ThreadId = thread
	c.send(request)
}

func (c *DAPClient) NextRequest(thread int) {
	request := &dap.NextRequest{Request: *c.newRequest("next")}
	request.Arguments.ThreadId = thread
	c.send(request)
}

func (c *DAPClient) NextInstructionRequest(thread int) {
	request := &dap.NextRequest{Request: *c.newRequest("next")}
	request.Arguments.ThreadId = thread
	request.Arguments.Granularity = "instruction"
	c.send(request)
}

func (c *DAPClient) StepInRequest(thread int) {
	request := &dap.StepInRequest{Request: *c.newRequest("stepIn")}
	request.Arguments.ThreadId = thread
	c.send(request)
}

func (c *DAPClient) StepInInstructionRequest(thread int) {
	request := &dap.StepInRequest{Request: *c.newRequest("stepIn")}
	request.Arguments.ThreadId = thread
	request.Arguments.Granularity = "instruction"
	c.send(request)
}

func (c *DAPClient) StepOutRequest(thread int) {
	request := &dap.StepOutRequest{Request: *c.newRequest("stepOut")}
	request.Arguments.ThreadId = thread
	c.send(request)
}

func (c *DAPClient) StepOutInstructionRequest(thread int) {
	request := &dap.StepOutRequest{Request: *c.newRequest("stepOut")}
	request.Arguments.ThreadId = thread
	request.Arguments.Granularity = "instruction"
	c.send(request)
}

func (c *DAPClient) PauseRequest(threadId int) {
	request := &dap.PauseRequest{Request: *c.newRequest("pause")}
	request.Arguments.ThreadId = threadId
	c.send(request)
}

func (c *DAPClient) ThreadsRequest() {
	request := &dap.ThreadsRequest{Request: *c.newRequest("threads")}
	c.send(request)
}

func (c *DAPClient) StackTraceRequest(threadID, startFrame, levels int) {
	request := &dap.StackTraceRequest{Request: *c.newRequest("stackTrace")}
	request.Arguments.ThreadId = threadID
	request.Arguments.StartFrame = startFrame
	request.Arguments.Levels = levels
	c.send(request)
}

func (c *DAPClient) ScopesRequest(frameID int) {
	request := &dap.ScopesRequest{Request: *c.newRequest("scopes")}
	request.Arguments.FrameId = frameID
	c.send(request)
}

func (c *DAPClient) VariablesRequest(variablesReference int) {
	request := &dap.VariablesRequest{Request: *c.newRequest("variables")}
	request.Arguments.VariablesReference = variablesReference
	c.send(request)
}

func (c *DAPClient) IndexedVariablesRequest(variablesReference, start, count int) {
	request := &dap.VariablesRequest{Request: *c.newRequest("variables")}
	request.Arguments.VariablesReference = variablesReference
	request.Arguments.Filter = "indexed"
	request.Arguments.Start = start
	request.Arguments.Count = count
	c.send(request)
}

func (c *DAPClient) NamedVariablesRequest(variablesReference int) {
	request := &dap.VariablesRequest{Request: *c.newRequest("variables")}
	request.Arguments.VariablesReference = variablesReference
	request.Arguments.Filter = "named"
	c.send(request)
}

func (c *DAPClient) TerminateRequest() {
	c.send(&dap.TerminateRequest{Request: *c.newRequest("terminate")})
}

func (c *DAPClient) RestartRequest() {
	c.send(&dap.RestartRequest{Request: *c.newRequest("restart")})
}

func (c *DAPClient) SetFunctionBreakpointsRequest(breakpoints []dap.FunctionBreakpoint) {
	c.send(&dap.SetFunctionBreakpointsRequest{
		Request: *c.newRequest("setFunctionBreakpoints"),
		Arguments: dap.SetFunctionBreakpointsArguments{
			Breakpoints: breakpoints,
		},
	})
}

func (c *DAPClient) SetInstructionBreakpointsRequest(breakpoints []dap.InstructionBreakpoint) {
	c.send(&dap.SetInstructionBreakpointsRequest{
		Request: *c.newRequest("setInstructionBreakpoints"),
		Arguments: dap.SetInstructionBreakpointsArguments{
			Breakpoints: breakpoints,
		},
	})
}

func (c *DAPClient) StepBackRequest() {
	c.send(&dap.StepBackRequest{Request: *c.newRequest("stepBack")})
}

func (c *DAPClient) ReverseContinueRequest() {
	c.send(&dap.ReverseContinueRequest{Request: *c.newRequest("reverseContinue")})
}

func (c *DAPClient) SetVariableRequest(variablesRef int, name, value string) {
	request := &dap.SetVariableRequest{Request: *c.newRequest("setVariable")}
	request.Arguments.VariablesReference = variablesRef
	request.Arguments.Name = name
	request.Arguments.Value = value
	c.send(request)
}

func (c *DAPClient) RestartFrameRequest() {
	c.send(&dap.RestartFrameRequest{Request: *c.newRequest("restartFrame")})
}

func (c *DAPClient) GotoRequest() {
	c.send(&dap.GotoRequest{Request: *c.newRequest("goto")})
}

func (c *DAPClient) SetExpressionRequest() {
	c.send(&dap.SetExpressionRequest{Request: *c.newRequest("setExpression")})
}

func (c *DAPClient) SourceRequest() {
	c.send(&dap.SourceRequest{Request: *c.newRequest("source")})
}

func (c *DAPClient) TerminateThreadsRequest() {
	c.send(&dap.TerminateThreadsRequest{Request: *c.newRequest("terminateThreads")})
}

func (c *DAPClient) EvaluateRequest(expr string, fid int, context string) {
	request := &dap.EvaluateRequest{Request: *c.newRequest("evaluate")}
	request.Arguments.Expression = expr
	request.Arguments.FrameId = fid
	request.Arguments.Context = context
	c.send(request)
}

func (c *DAPClient) StepInTargetsRequest() {
	c.send(&dap.StepInTargetsRequest{Request: *c.newRequest("stepInTargets")})
}

func (c *DAPClient) GotoTargetsRequest() {
	c.send(&dap.GotoTargetsRequest{Request: *c.newRequest("gotoTargets")})
}

func (c *DAPClient) CompletionsRequest() {
	c.send(&dap.CompletionsRequest{Request: *c.newRequest("completions")})
}

func (c *DAPClient) ExceptionInfoRequest(threadID int) {
	request := &dap.ExceptionInfoRequest{Request: *c.newRequest("exceptionInfo")}
	request.Arguments.ThreadId = threadID
	c.send(request)
}

func (c *DAPClient) LoadedSourcesRequest() {
	c.send(&dap.LoadedSourcesRequest{Request: *c.newRequest("loadedSources")})
}

func (c *DAPClient) DataBreakpointInfoRequest() {
	c.send(&dap.DataBreakpointInfoRequest{Request: *c.newRequest("dataBreakpointInfo")})
}

func (c *DAPClient) SetDataBreakpointsRequest() {
	c.send(&dap.SetDataBreakpointsRequest{Request: *c.newRequest("setDataBreakpoints")})
}

func (c *DAPClient) ReadMemoryRequest() {
	c.send(&dap.ReadMemoryRequest{Request: *c.newRequest("readMemory")})
}

func (c *DAPClient) DisassembleRequest(memoryReference string, instructionOffset, inctructionCount int) {
	c.send(&dap.DisassembleRequest{
		Request: *c.newRequest("disassemble"),
		Arguments: dap.DisassembleArguments{
			MemoryReference:   memoryReference,
			Offset:            0,
			InstructionOffset: instructionOffset,
			InstructionCount:  inctructionCount,
			ResolveSymbols:    false,
		},
	})
}

func (c *DAPClient) CancelRequest() {
	c.send(&dap.CancelRequest{Request: *c.newRequest("cancel")})
}

func (c *DAPClient) BreakpointLocationsRequest() {
	c.send(&dap.BreakpointLocationsRequest{Request: *c.newRequest("breakpointLocations")})
}

func (c *DAPClient) ModulesRequest() {
	c.send(&dap.ModulesRequest{Request: *c.newRequest("modules")})
}

func (c *DAPClient) newRequest(command string) *dap.Request {
	request := &dap.Request{}
	request.Type = "request"
	request.Command = command
	request.Seq = c.seq
	c.seq++
	return request
}
