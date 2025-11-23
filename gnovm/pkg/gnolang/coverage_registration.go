package gnolang

import (
)

// RegisterCoverageBlocks walks the AST and registers coverage blocks for all statements
func (m *Machine) RegisterCoverageBlocks(fset *FileSet) {
	if m.Coverage == nil {
		return
	}

	// Deduplicate files to avoid registering blocks multiple times
	// This can happen if RunMemPackage is called multiple times on the same PackageNode
	// which accumulates files in its FileSet.
	seenFiles := make(map[string]bool)
	for _, fn := range fset.Files {
		if seenFiles[fn.FileName] {
			continue
		}
		seenFiles[fn.FileName] = true

		// Use Transcribe to walk the AST
		Transcribe(fn, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
			if stage != TRANS_ENTER {
				return n, TRANS_CONTINUE
			}

			// Check if it's a statement
			if stmt, ok := n.(Stmt); ok {
				// Skip empty statements or blocks that are just containers
				if _, ok := stmt.(*BlockStmt); ok {
					return n, TRANS_CONTINUE
				}
				if _, ok := stmt.(*EmptyStmt); ok {
					return n, TRANS_CONTINUE
				}

				// Get span info
				span := stmt.GetSpan()
				pos := span.Pos
				if pos.Line == 0 {
					return n, TRANS_CONTINUE
				}

				endLine := span.End.Line
				endCol := span.End.Column
				if endLine == 0 {
					endLine = pos.Line
				}
				if endCol == 0 {
					endCol = pos.Column + 10 // Rough estimate if missing
				}

				// Register the block
				// Use full package path for filename to be compatible with go tool cover
				fullPath := m.Coverage.PkgPath + "/" + fn.FileName
				// Register block for coverage tracking
				idx := m.Coverage.RegisterBlock(
					fullPath,
					pos.Line,
					pos.Column,
					endLine,
					endCol,
					1, // 1 statement
				)

				// Store the index on the node
				stmt.SetCoverageBlockIndex(idx)
			}

			return n, TRANS_CONTINUE
		})
	}
	// Coverage blocks registered successfully
}
