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

// handleAttach processes the attach request
func (s *DAPServer) handleAttach(req *Request, _ []byte) error {
	var args AttachArguments
	if err := json.Unmarshal(req.Arguments, &args); err != nil {
		return err
	}

	// For Gno debugger, attach mode means the program was started with --attach flag
	// and is waiting for debugger to connect before starting execution.

	// If we're in attach mode, start the program execution now
	if s.attachMode && s.programFiles != nil {
		go func() {
			// Run the loaded files
			s.machine.RunFiles(s.programFiles...)
			// Run main expression
			if s.mainExpr != "" {
				ex, err := ParseExpr(s.mainExpr)
				if err == nil {
					s.machine.Eval(ex)
				}
			}

			// Send terminated event when done
			s.SendTerminatedEvent()
		}()
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
	if len(s.debugger.breakpoints) == 0 {
		// No breakpoints set, continue execution
		s.debugger.state = DebugAtRun
	} else {
		// Breakpoints exist, wait for continue command
		s.debugger.state = DebugAtCmd
	}
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
	var args VariablesArguments
	if err := json.Unmarshal(req.Arguments, &args); err != nil {
		return err
	}

	variables := make([]Variable, 0)

	// Determine scope from variablesReference
	// We use a simple scheme: 1000+frameID for locals, 2000+frameID for globals, 3000+ for expandable variables
	if args.VariablesReference >= 1000 && args.VariablesReference < 2000 {
		// Local variables
		frameID := args.VariablesReference - 1000
		variables = s.getLocalVariables(frameID)
	} else if args.VariablesReference >= 2000 && args.VariablesReference < 3000 {
		// Global variables
		variables = s.getGlobalVariables()
	} else if args.VariablesReference >= 3000 {
		// Expandable variable
		variables = s.getExpandedVariables(args.VariablesReference)
	}

	resp := NewResponse(req, true)
	resp.Body = map[string]any{
		"variables": variables,
	}
	return s.sendMessage(resp)
}

// getLocalVariables returns local variables for a specific frame
func (s *DAPServer) getLocalVariables(frameID int) []Variable {
	variables := make([]Variable, 0)

	// Get variables from the current frame
	if len(s.machine.Blocks) == 0 {
		return variables
	}

	// Save current frame level
	oldFrameLevel := s.debugger.frameLevel
	s.debugger.frameLevel = frameID

	// Collect all variable names from visible blocks
	varMap := make(map[string]TypedValue)

	// Iterate through blocks to find variables
	for i := len(s.machine.Blocks) - 1; i >= 0; i-- {
		block := s.machine.Blocks[i]
		if block == nil || block.Source == nil {
			continue
		}

		// Get block names
		names := block.Source.GetBlockNames()
		for idx, name := range names {
			nameStr := string(name)
			// Skip if we already have this variable (inner scope shadows outer)
			if _, exists := varMap[nameStr]; exists {
				continue
			}
			if idx < len(block.Values) {
				varMap[nameStr] = block.Values[idx]
			}
		}

		// Stop at function boundary
		if i > 0 && s.machine.Frames != nil {
			for _, frame := range s.machine.Frames {
				if frame.Func != nil && block.Source == frame.Func.Source {
					goto done
				}
			}
		}
	}

done:
	// Convert to Variable array
	for name, tv := range varMap {
		variables = append(variables, s.typedValueToVariable(name, tv))
	}

	// Restore frame level
	s.debugger.frameLevel = oldFrameLevel

	return variables
}

// getGlobalVariables returns global/package-level variables
func (s *DAPServer) getGlobalVariables() []Variable {
	variables := make([]Variable, 0)

	// Get global block (usually first block)
	if len(s.machine.Blocks) > 0 && s.machine.Blocks[0] != nil {
		block := s.machine.Blocks[0]
		if block.Source != nil {
			names := block.Source.GetBlockNames()
			for idx, name := range names {
				if idx < len(block.Values) {
					variables = append(variables, s.typedValueToVariable(string(name), block.Values[idx]))
				}
			}
		}
	}

	return variables
}

// typedValueToVariable converts a TypedValue to a DAP Variable
func (s *DAPServer) typedValueToVariable(name string, tv TypedValue) Variable {
	var valueStr, typeStr string
	variablesRef := 0

	if tv.T != nil {
		typeStr = tv.T.String()
	}

	// Format value based on type
	switch t := tv.T.(type) {
	case PrimitiveType:
		valueStr = tv.String()
	case *SliceType:
		if tv.V != nil {
			switch v := tv.V.(type) {
			case *SliceValue:
				valueStr = fmt.Sprintf("[]%s (length=%d, cap=%d)", t.Elt.String(), v.Length, v.Maxcap)
				if v.Length > 0 {
					// Assign a unique reference for expanding the slice
					variablesRef = s.assignVariableRef(tv, 0)
				}
			default:
				valueStr = tv.String()
			}
		} else {
			valueStr = "nil"
		}
	case *ArrayType:
		if tv.V != nil {
			switch v := tv.V.(type) {
			case *ArrayValue:
				length := 0
				if v != nil && v.List != nil {
					length = len(v.List)
				}
				valueStr = fmt.Sprintf("[%d]%s", t.Len, t.Elt.String())
				if length > 0 {
					// Assign a unique reference for expanding the array
					variablesRef = s.assignVariableRef(tv, 0)
				}
			default:
				valueStr = tv.String()
			}
		} else {
			valueStr = "nil"
		}
	case *StructType:
		if tv.V != nil {
			if sv, ok := tv.V.(*StructValue); ok && sv != nil {
				valueStr = t.String()
				// Assign a unique reference for expanding the struct
				variablesRef = s.assignVariableRef(tv, 0)
			} else {
				valueStr = tv.String()
			}
		} else {
			valueStr = "nil"
		}
	case *MapType:
		valueStr = t.String()
		if tv.V != nil {
			if mv, ok := tv.V.(*MapValue); ok && mv != nil {
				length := 0
				if mv.List != nil {
					length = mv.List.Size
				}
				valueStr = fmt.Sprintf("map[%s]%s (length=%d)", t.Key.String(), t.Value.String(), length)
				if length > 0 {
					// Assign a unique reference for expanding the map
					variablesRef = s.assignVariableRef(tv, 0)
				}
			}
		}
	case *PointerType:
		if tv.V != nil {
			if pv, ok := tv.V.(PointerValue); ok && pv.TV != nil {
				valueStr = fmt.Sprintf("&%s", pv.TV.String())
				// Assign a unique reference for dereferencing the pointer
				variablesRef = s.assignVariableRef(*pv.TV, 0)
			} else {
				valueStr = tv.String()
			}
		} else {
			valueStr = "nil"
		}
	case *DeclaredType:
		// Handle declared types (custom types)
		valueStr = tv.String()
		if tv.V != nil {
			// Check the underlying type
			if t.Base != nil {
				switch t.Base.(type) {
				case *StructType:
					if sv, ok := tv.V.(*StructValue); ok && sv != nil {
						// Assign a unique reference for expanding the struct
						variablesRef = s.assignVariableRef(tv, 0)
					}
				case *SliceType:
					if sv, ok := tv.V.(*SliceValue); ok && sv != nil && sv.Length > 0 {
						// Assign a unique reference for expanding the slice
						variablesRef = s.assignVariableRef(tv, 0)
					}
				case *ArrayType:
					if av, ok := tv.V.(*ArrayValue); ok && av != nil && av.List != nil && len(av.List) > 0 {
						// Assign a unique reference for expanding the array
						variablesRef = s.assignVariableRef(tv, 0)
					}
				}
			}
		}
	default:
		// Check if it's a heapitem
		if tv.T != nil && tv.T.Kind() == HeapItemKind {
			// Handle heapitem by unwrapping the actual value
			if hiv, ok := tv.V.(*HeapItemValue); ok && hiv != nil {
				// Recursively process the wrapped value
				return s.typedValueToVariable(name, hiv.Value)
			}
		}
		// For other types, use String() method
		valueStr = tv.String()
	}

	return Variable{
		Name:               name,
		Value:              valueStr,
		Type:               typeStr,
		VariablesReference: variablesRef,
	}
}

// assignVariableRef assigns a unique reference ID for a variable that can be expanded
func (s *DAPServer) assignVariableRef(tv TypedValue, frameID int) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	ref := s.nextVariableRef
	s.nextVariableRef++
	s.variableRefs[ref] = variableInfo{
		value:    tv,
		frameID:  frameID,
		isScoped: false,
	}
	return ref
}

// getExpandedVariables returns child variables for an expandable variable
func (s *DAPServer) getExpandedVariables(ref int) []Variable {
	s.mu.Lock()
	info, exists := s.variableRefs[ref]
	s.mu.Unlock()

	if !exists {
		return []Variable{}
	}

	variables := make([]Variable, 0)
	tv := info.value

	switch t := tv.T.(type) {
	case *ArrayType:
		if av, ok := tv.V.(*ArrayValue); ok && av != nil && av.List != nil {
			for i, elem := range av.List {
				name := fmt.Sprintf("[%d]", i)
				variables = append(variables, s.typedValueToVariable(name, elem))
			}
		}
	case *SliceType:
		if sv, ok := tv.V.(*SliceValue); ok && sv != nil {
			// Get the underlying array
			if base := sv.GetBase(s.machine.Store); base != nil && base.List != nil {
				for i := sv.Offset; i < sv.Offset+sv.Length && i < len(base.List); i++ {
					name := fmt.Sprintf("[%d]", i-sv.Offset)
					variables = append(variables, s.typedValueToVariable(name, base.List[i]))
				}
			}
		}
	case *StructType:
		if sv, ok := tv.V.(*StructValue); ok && sv != nil && sv.Fields != nil {
			for i, field := range t.Fields {
				if i < len(sv.Fields) {
					variables = append(variables, s.typedValueToVariable(string(field.Name), sv.Fields[i]))
				}
			}
		}
	case *MapType:
		if mv, ok := tv.V.(*MapValue); ok && mv != nil && mv.List != nil {
			i := 0
			for item := mv.List.Head; item != nil; item = item.Next {
				keyVar := s.typedValueToVariable(fmt.Sprintf("[key %d]", i), item.Key)
				variables = append(variables, keyVar)

				valueVar := s.typedValueToVariable(fmt.Sprintf("[value %d]", i), item.Value)
				variables = append(variables, valueVar)

				i++
			}
		}
	case *PointerType:
		// For pointers, we've already stored the dereferenced value
		// Just create a single variable for the pointed-to value
		variables = append(variables, s.typedValueToVariable("*", tv))
	case *DeclaredType:
		// Handle declared types by checking their base type
		if t.Base != nil {
			switch baseType := t.Base.(type) {
			case *StructType:
				if sv, ok := tv.V.(*StructValue); ok && sv != nil && sv.Fields != nil {
					for i, field := range baseType.Fields {
						if i < len(sv.Fields) {
							variables = append(variables, s.typedValueToVariable(string(field.Name), sv.Fields[i]))
						}
					}
				}
			case *SliceType:
				// Reuse slice expansion logic
				if sv, ok := tv.V.(*SliceValue); ok && sv != nil {
					if base := sv.GetBase(s.machine.Store); base != nil && base.List != nil {
						for i := sv.Offset; i < sv.Offset+sv.Length && i < len(base.List); i++ {
							name := fmt.Sprintf("[%d]", i-sv.Offset)
							variables = append(variables, s.typedValueToVariable(name, base.List[i]))
						}
					}
				}
			case *ArrayType:
				// Reuse array expansion logic
				if av, ok := tv.V.(*ArrayValue); ok && av != nil && av.List != nil {
					for i, elem := range av.List {
						name := fmt.Sprintf("[%d]", i)
						variables = append(variables, s.typedValueToVariable(name, elem))
					}
				}
			}
		}
	}

	return variables
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
	// Check if DAP server is ready to send messages
	if s.writer == nil {
		// DAP server not yet connected, skip sending event
		return nil
	}

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
