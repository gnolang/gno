package gnolang

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// handleInitialize processes the initialize request
func (s *DAPServer) handleInitialize(req *Request, _ []byte) error {
	var args InitializeArguments
	if err := json.Unmarshal(req.Arguments, &args); err != nil {
		return err
	}

	// Store client capabilities
	s.clientLinesStartAt1 = args.LinesStartAt1
	s.clientColumnsStartAt1 = args.ColumnsStartAt1
	s.clientPathFormat = args.PathFormat

	// Create response with our capabilities
	resp := NewResponse(req, true)
	resp.Body = Capabilities{
		SupportsConfigurationDoneRequest:   true,
		SupportsConditionalBreakpoints:     false,
		SupportsHitConditionalBreakpoints:  false,
		SupportsEvaluateForHovers:          true,
		SupportsStepBack:                   false,
		SupportsSetVariable:                false,
		SupportsRestartFrame:               false,
		SupportsGotoTargetsRequest:         false,
		SupportsStepInTargetsRequest:       false,
		SupportsCompletionsRequest:         false,
		SupportsModulesRequest:             false,
		SupportsRestartRequest:             false,
		SupportsExceptionOptions:           false,
		SupportsValueFormattingOptions:     false,
		SupportsExceptionInfoRequest:       false,
		SupportTerminateDebuggee:           true,
		SupportsDelayedStackTraceLoading:   false,
		SupportsLoadedSourcesRequest:       false,
		SupportsLogPoints:                  false,
		SupportsTerminateThreadsRequest:    false,
		SupportsSetExpression:              false,
		SupportsTerminateRequest:           true,
		SupportsDataBreakpoints:            false,
		SupportsReadMemoryRequest:          false,
		SupportsDisassembleRequest:         false,
		SupportsCancelRequest:              false,
		SupportsBreakpointLocationsRequest: false,
	}

	s.initialized = true
	return s.sendMessage(resp)
}

// handleLaunch processes the launch request
func (s *DAPServer) handleLaunch(req *Request, _ []byte) error {
	var args LaunchArguments
	if err := json.Unmarshal(req.Arguments, &args); err != nil {
		return err
	}

	// Send response
	resp := NewResponse(req, true)
	if err := s.sendMessage(resp); err != nil {
		return err
	}

	// Send initialized event
	event := &InitializedEvent{
		Event: *NewEvent("initialized"),
	}
	return s.sendMessage(event)
}

// handleSetBreakpoints processes the setBreakpoints request
func (s *DAPServer) handleSetBreakpoints(req *Request, _ []byte) error {
	var args SetBreakpointsArguments
	if err := json.Unmarshal(req.Arguments, &args); err != nil {
		return err
	}

	// Clear existing breakpoints for this source
	delete(s.breakpoints, args.Source.Path)

	// Set new breakpoints
	var responseBreakpoints []Breakpoint
	for _, bp := range args.Breakpoints {
		line := s.convertClientToServerLine(bp.Line)

		// Handle column: if not specified (0), keep it as 0
		column := 0
		if bp.Column > 0 {
			column = s.convertClientToServerColumn(bp.Column)
		}

		// Create a location for the breakpoint
		loc := Location{
			File: filepath.Base(args.Source.Path),
			Span: Span{
				Pos: Pos{Line: line, Column: column},
			},
		}

		// Add to debugger's breakpoint list
		s.debugger.breakpoints = append(s.debugger.breakpoints, loc)

		// Create response breakpoint
		respBp := Breakpoint{
			ID:       s.nextBreakpointID,
			Verified: true,
			Source:   args.Source,
			Line:     bp.Line,
			Column:   bp.Column,
		}
		s.nextBreakpointID++

		responseBreakpoints = append(responseBreakpoints, respBp)
	}

	s.breakpoints[args.Source.Path] = responseBreakpoints

	// Send response
	resp := NewResponse(req, true)
	resp.Body = map[string]any{
		"breakpoints": responseBreakpoints,
	}
	return s.sendMessage(resp)
}

// handleConfigurationDone processes the configurationDone request
func (s *DAPServer) handleConfigurationDone(req *Request) error {
	resp := NewResponse(req, true)
	if err := s.sendMessage(resp); err != nil {
		return err
	}

	// Start debugging
	s.debugger.state = DebugAtCmd
	return nil
}

// handleThreads processes the threads request
func (s *DAPServer) handleThreads(req *Request) error {
	resp := NewResponse(req, true)
	resp.Body = map[string]any{
		"threads": []map[string]any{
			{
				"id":   s.threadID,
				"name": "main",
			},
		},
	}
	return s.sendMessage(resp)
}

// handleStackTrace processes the stackTrace request
func (s *DAPServer) handleStackTrace(req *Request, _ []byte) error {
	var args StackTraceArguments
	if err := json.Unmarshal(req.Arguments, &args); err != nil {
		return err
	}

	// Build stack frames from machine state
	var frames []StackFrame
	frameIndex := 0

	// Add current frame
	loc := s.debugger.loc
	if loc.File != "" {
		frame := StackFrame{
			ID:     frameIndex,
			Name:   s.getCurrentFunctionName(),
			Line:   s.convertServerToClientLine(loc.Line),
			Column: s.convertServerToClientColumn(loc.Column),
		}

		// Convert path if needed
		if loc.PkgPath != "" {
			frame.Source = Source{
				Name: filepath.Base(loc.File),
				Path: filepath.Join(loc.PkgPath, loc.File),
			}
		} else {
			frame.Source = Source{
				Name: filepath.Base(loc.File),
				Path: loc.File,
			}
		}

		frames = append(frames, frame)
		frameIndex++
	}

	// Add frames from call stack
	for i := len(s.debugger.call) - 1; i >= 0 && frameIndex < args.Levels; i-- {
		callLoc := s.debugger.call[i]
		frame := StackFrame{
			ID:     frameIndex,
			Name:   fmt.Sprintf("frame_%d", frameIndex),
			Line:   s.convertServerToClientLine(callLoc.Line),
			Column: s.convertServerToClientColumn(callLoc.Column),
		}

		if callLoc.PkgPath != "" {
			frame.Source = Source{
				Name: filepath.Base(callLoc.File),
				Path: filepath.Join(callLoc.PkgPath, callLoc.File),
			}
		} else {
			frame.Source = Source{
				Name: filepath.Base(callLoc.File),
				Path: callLoc.File,
			}
		}

		frames = append(frames, frame)
		frameIndex++
	}

	resp := NewResponse(req, true)
	resp.Body = map[string]any{
		"stackFrames": frames,
		"totalFrames": len(frames),
	}
	return s.sendMessage(resp)
}

// handleScopes processes the scopes request
func (s *DAPServer) handleScopes(req *Request, _ []byte) error {
	var args struct {
		FrameID int `json:"frameId"`
	}
	if err := json.Unmarshal(req.Arguments, &args); err != nil {
		return err
	}

	// For now, we'll just return local and global scopes
	resp := NewResponse(req, true)
	resp.Body = map[string]any{
		"scopes": []map[string]any{
			{
				"name":               "Locals",
				"variablesReference": 1000 + args.FrameID, // Unique reference
				"expensive":          false,
			},
			{
				"name":               "Globals",
				"variablesReference": 2000 + args.FrameID, // Unique reference
				"expensive":          false,
			},
		},
	}
	return s.sendMessage(resp)
}

// handleVariables processes the variables request
func (s *DAPServer) handleVariables(req *Request, _ []byte) error {
	var args struct {
		VariablesReference int `json:"variablesReference"`
	}
	if err := json.Unmarshal(req.Arguments, &args); err != nil {
		return err
	}

	// TODO: Implement actual variable lookup
	// For now, return empty list
	resp := NewResponse(req, true)
	resp.Body = map[string]any{
		"variables": []map[string]any{},
	}
	return s.sendMessage(resp)
}

// handleContinue processes the continue request
func (s *DAPServer) handleContinue(req *Request, _ []byte) error {
	var args ContinueArguments
	if err := json.Unmarshal(req.Arguments, &args); err != nil {
		return err
	}

	// Send response first
	resp := NewResponse(req, true)
	resp.Body = map[string]any{
		"allThreadsContinued": true,
	}
	if err := s.sendMessage(resp); err != nil {
		return err
	}

	// Set debugger to continue
	s.debugger.lastCmd = "continue"
	s.debugger.state = DebugAtRun

	// Send continued event
	event := &ContinuedEvent{
		Event: *NewEvent("continued"),
		Body: ContinuedEventBody{
			ThreadID:            s.threadID,
			AllThreadsContinued: true,
		},
	}
	return s.sendMessage(event)
}

// handleNext processes the next request
func (s *DAPServer) handleNext(req *Request) error {
	// Send response
	resp := NewResponse(req, true)
	if err := s.sendMessage(resp); err != nil {
		return err
	}

	// Set debugger to step over
	s.debugger.lastCmd = "next"
	s.debugger.state = DebugAtRun
	s.debugger.nextDepth = callDepth(s.machine)
	s.debugger.nextLoc = s.debugger.loc

	return nil
}

// handleStepIn processes the stepIn request
func (s *DAPServer) handleStepIn(req *Request) error {
	// Send response
	resp := NewResponse(req, true)
	if err := s.sendMessage(resp); err != nil {
		return err
	}

	// Set debugger to step
	s.debugger.lastCmd = "step"
	s.debugger.state = DebugAtRun

	return nil
}

// handleStepOut processes the stepOut request
func (s *DAPServer) handleStepOut(req *Request) error {
	// Send response
	resp := NewResponse(req, true)
	if err := s.sendMessage(resp); err != nil {
		return err
	}

	// Set debugger to step out
	s.debugger.lastCmd = "stepout"
	s.debugger.state = DebugAtRun
	s.debugger.nextDepth = callDepth(s.machine)

	return nil
}

// handlePause processes the pause request
func (s *DAPServer) handlePause(req *Request) error {
	// Send response
	resp := NewResponse(req, true)
	if err := s.sendMessage(resp); err != nil {
		return err
	}

	// Force debugger to pause at next instruction
	s.debugger.state = DebugAtCmd

	return nil
}

// handleEvaluate processes the evaluate request
func (s *DAPServer) handleEvaluate(req *Request, _ []byte) error {
	var args EvaluateArguments
	if err := json.Unmarshal(req.Arguments, &args); err != nil {
		return err
	}

	// Use existing debugPrint functionality
	result := ""
	err := func() error {
		// Capture output
		oldOut := s.machine.Output
		var buf strings.Builder
		s.machine.Output = &buf
		defer func() { s.machine.Output = oldOut }()

		// Evaluate expression
		if err := debugPrint(s.machine, args.Expression); err != nil {
			return err
		}

		result = strings.TrimSpace(buf.String())
		return nil
	}()

	resp := NewResponse(req, err == nil)
	if err != nil {
		resp.Message = err.Error()
	} else {
		resp.Body = map[string]any{
			"result":             result,
			"variablesReference": 0,
		}
	}
	return s.sendMessage(resp)
}

// handleDisconnect processes the disconnect request
func (s *DAPServer) handleDisconnect(req *Request) error {
	// Send response
	resp := NewResponse(req, true)
	if err := s.sendMessage(resp); err != nil {
		return err
	}

	// Terminate debugging
	s.terminated = true
	s.debugger.enabled = false
	s.debugger.state = DebugAtExit

	// Send terminated event
	event := &TerminatedEvent{
		Event: *NewEvent("terminated"),
	}
	return s.sendMessage(event)
}

// getCurrentFunctionName returns the name of the current function
func (s *DAPServer) getCurrentFunctionName() string {
	if s.machine.Package != nil {
		name := string(s.machine.Package.PkgName)
		if len(s.machine.Frames) > 0 {
			f := s.machine.Frames[len(s.machine.Frames)-1]
			if f.Func != nil {
				name += "." + string(f.Func.Name) + "()"
			}
		}
		return name
	}
	return "main"
}

// SendStoppedEvent sends a stopped event when execution pauses
func (s *DAPServer) SendStoppedEvent(reason string, description string) error {
	event := &StoppedEvent{
		Event: *NewEvent("stopped"),
		Body: StoppedEventBody{
			Reason:            reason,
			Description:       description,
			ThreadID:          s.threadID,
			AllThreadsStopped: true,
		},
	}
	return s.sendMessage(event)
}
