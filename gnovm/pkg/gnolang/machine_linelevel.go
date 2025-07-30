package gnolang

// getCurrentLocation extracts the current execution location from the machine state
func (m *Machine) getCurrentLocation() *profileLocation {
	if m == nil {
		return nil
	}

	// Try to get location from current statement
	if len(m.Stmts) > 0 {
		stmt := m.PeekStmt(1)
		if stmt != nil {
			loc := &profileLocation{
				line:   stmt.GetLine(),
				column: stmt.GetColumn(),
			}

			// Get file information from current package
			if m.Package != nil {
				loc.file = m.Package.PkgPath
			}

			// Get function information from current frame
			if len(m.Frames) > 0 {
				frame := &m.Frames[len(m.Frames)-1]
				if frame.Func != nil {
					loc.function = string(frame.Func.Name)
					if frame.Func.PkgPath != "" {
						loc.function = frame.Func.PkgPath + "." + loc.function
					}
				}
			}

			return loc
		}
	}

	// Try to get location from current expression
	if len(m.Exprs) > 0 {
		expr := m.PeekExpr(1)
		if expr != nil {
			loc := &profileLocation{
				line:   expr.GetLine(),
				column: expr.GetColumn(),
			}

			// Get file information
			if m.Package != nil {
				loc.file = m.Package.PkgPath
			}

			// Get function information from current frame
			if len(m.Frames) > 0 {
				frame := &m.Frames[len(m.Frames)-1]
				if frame.Func != nil {
					loc.function = string(frame.Func.Name)
					if frame.Func.PkgPath != "" {
						loc.function = frame.Func.PkgPath + "." + loc.function
					}
				}
			}

			return loc
		}
	}

	// Fall back to frame source if available
	if len(m.Frames) > 0 {
		frame := &m.Frames[len(m.Frames)-1]
		if frame.Source != nil {
			loc := &profileLocation{
				line:   frame.Source.GetLine(),
				column: frame.Source.GetColumn(),
			}

			// Get file information
			if m.Package != nil {
				loc.file = m.Package.PkgPath
			}

			// Get function information
			if frame.Func != nil {
				loc.function = string(frame.Func.Name)
				if frame.Func.PkgPath != "" {
					loc.function = frame.Func.PkgPath + "." + loc.function
				}
			}

			return loc
		}
	}

	return nil
}

// RecordCurrentLocation records profiling data for the current execution location
func (m *Machine) RecordCurrentLocation(cycles int64) {
	if m.Profiler == nil || !m.Profiler.enabled || !m.Profiler.lineLevel {
		return
	}

	loc := m.getCurrentLocation()
	if loc != nil {
		m.Profiler.RecordLineLevel(m, loc, cycles)
	}
}

// Enhanced Run method hook for line-level profiling
func (m *Machine) recordOpLocation(op Op, cycles int64) {
	if m.IsProfilingEnabled() && m.Profiler.lineLevel {
		// Only record for significant operations
		switch op {
		case OpCall, OpEval, OpExec, OpAssign, OpAddAssign, OpSubAssign,
			OpMulAssign, OpQuoAssign, OpRemAssign, OpBandAssign,
			OpBorAssign, OpXorAssign, OpShlAssign, OpShrAssign,
			OpBandnAssign, OpDefine:
			m.RecordCurrentLocation(cycles)
		}
	}
}
