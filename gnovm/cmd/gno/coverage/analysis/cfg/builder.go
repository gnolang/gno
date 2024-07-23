package cfg

import (
	"go/ast"
	"go/token"
)

// ref: https://github.com/godoctor/godoctor/blob/master/analysis/cfg/cfg.go

type builder struct {
	blocks      map[ast.Stmt]*block
	prev        []ast.Stmt        // blocks to hook up to current block
	branches    []*ast.BranchStmt // accumulated branches from current inner blocks
	entry, exit *ast.BadStmt      // single-entry, single-exit nodes
	defers      []*ast.DeferStmt  // all defers encountered
}

// NewBuilder constructs a CFG from the given slice of statements.
func NewBuilder() *builder {
	// The ENTRY and EXIT nodes are given positions -2 and -1 so cfg.Sort
	// will work correct: ENTRY will always be first, followed by EXIT,
	// followed by the other CFG nodes.
	return &builder{
		blocks: map[ast.Stmt]*block{},
		entry:  &ast.BadStmt{From: -2, To: -2},
		exit:   &ast.BadStmt{From: -1, To: -1},
	}
}

// Build runs buildBlock on the given block (traversing nested statements), and
// adds entry and exit nodes.
func (b *builder) Build(s []ast.Stmt) *CFG {
	b.prev = []ast.Stmt{b.entry}
	b.buildBlock(s)
	b.addSucc(b.exit)

	return &CFG{
		blocks: b.blocks,
		Entry:  b.entry,
		Exit:   b.exit,
		Defers: b.defers,
	}
}

// addSucc adds a control flow edge from all previous blocks to the block for the given statement.
// It updates both the successors of the previous blocks and the predecessors of the current block.
func (b *builder) addSucc(current ast.Stmt) {
	cur := b.block(current)

	for _, p := range b.prev {
		p := b.block(p)
		p.succs = appendNoDuplicates(p.succs, cur.stmt)
		cur.preds = appendNoDuplicates(cur.preds, p.stmt)
	}
}

func appendNoDuplicates(list []ast.Stmt, stmt ast.Stmt) []ast.Stmt {
	for _, s := range list {
		if s == stmt {
			return list
		}
	}
	return append(list, stmt)
}

// block returns a block for the given statement, creating one and inserting it
// into the CFG if it doesn't already exist.
func (b *builder) block(s ast.Stmt) *block {
	bl, ok := b.blocks[s]
	if !ok {
		bl = &block{stmt: s}
		b.blocks[s] = bl
	}
	return bl
}

// buildStmt adds the given statement and all nested statements to the control
// flow graph under construction. Upon completion, b.prev is set to all
// control flow exits generated from traversing cur.
func (b *builder) buildStmt(cur ast.Stmt) {
	if dfr, ok := cur.(*ast.DeferStmt); ok {
		b.defers = append(b.defers, dfr)
		return // never flow to or from defer
	}

	// Each buildXxx method will flow the previous blocks to itself appropriately and also
	// set the appropriate blocks to flow from at the end of the method.
	switch cur := cur.(type) {
	case *ast.BlockStmt:
		b.buildBlock(cur.List)
	case *ast.IfStmt:
		b.buildIf(cur)
	case *ast.ForStmt, *ast.RangeStmt:
		b.buildLoop(cur)
	case *ast.SwitchStmt, *ast.SelectStmt, *ast.TypeSwitchStmt:
		b.buildSwitch(cur)
	case *ast.BranchStmt:
		b.buildBranch(cur)
	case *ast.LabeledStmt:
		b.addSucc(cur)
		b.prev = []ast.Stmt{cur}
		b.buildStmt(cur.Stmt)
	case *ast.ReturnStmt:
		b.addSucc(cur)
		b.prev = []ast.Stmt{cur}
		b.addSucc(b.exit)
		b.prev = nil
	default: // most statements have straight-line control flow
		b.addSucc(cur)
		b.prev = []ast.Stmt{cur}
	}
}

// buildBranch handles the creation of CFG nodes for branch statements (break, continue, goto, fallthrough).
// It updates the CFG based on the type of branch statement.
func (b *builder) buildBranch(br *ast.BranchStmt) {
	b.addSucc(br)
	b.prev = []ast.Stmt{br}

	switch br.Tok {
	case token.FALLTHROUGH:
		// successors handled in buildSwitch, so skip this here
	case token.GOTO:
		b.addSucc(br.Label.Obj.Decl.(ast.Stmt)) // flow to label
	case token.BREAK, token.CONTINUE:
		b.branches = append(b.branches, br) // to handle at switch/for/etc level
	}
	b.prev = nil // successors handled elsewhere
}

// buildBlock iterates over a slice of statements, typically from an ast.BlockStmt,
// adding them successively to the CFG.
func (b *builder) buildBlock(block []ast.Stmt) {
	for _, stmt := range block {
		b.buildStmt(stmt)
	}
}

// buildIf constructs the CFG for an if statement, including its condition, body, and else clause (if present).
func (b *builder) buildIf(f *ast.IfStmt) {
	if f.Init != nil {
		b.addSucc(f.Init)
		b.prev = []ast.Stmt{f.Init}
	}
	b.addSucc(f)

	b.prev = []ast.Stmt{f}
	b.buildBlock(f.Body.List) // build then

	ctrlExits := b.prev // aggregate of b.prev from each condition

	switch s := f.Else.(type) {
	case *ast.BlockStmt: // build else
		b.prev = []ast.Stmt{f}
		b.buildBlock(s.List)
		ctrlExits = append(ctrlExits, b.prev...)
	case *ast.IfStmt: // build else if
		b.prev = []ast.Stmt{f}
		b.addSucc(s)
		b.buildIf(s)
		ctrlExits = append(ctrlExits, b.prev...)
	case nil: // no else
		ctrlExits = append(ctrlExits, f)
	}

	b.prev = ctrlExits
}

// buildLoop constructs the CFG for loop statements (for and range).
// It handles the initialization, condition, post-statement (for for loops), and body of the loop.
func (b *builder) buildLoop(stmt ast.Stmt) {
	// flows as such (range same w/o init & post):
	// previous -> [ init -> ] for -> body -> [ post -> ] for -> next

	var post ast.Stmt = stmt // post in for loop, or for stmt itself; body flows to this

	switch stmt := stmt.(type) {
	case *ast.ForStmt:
		if stmt.Init != nil {
			b.addSucc(stmt.Init)
			b.prev = []ast.Stmt{stmt.Init}
		}
		b.addSucc(stmt)

		if stmt.Post != nil {
			post = stmt.Post
			b.prev = []ast.Stmt{post}
			b.addSucc(stmt)
		}

		b.prev = []ast.Stmt{stmt}
		b.buildBlock(stmt.Body.List)
	case *ast.RangeStmt:
		b.addSucc(stmt)
		b.prev = []ast.Stmt{stmt}
		b.buildBlock(stmt.Body.List)
	}

	b.addSucc(post)

	ctrlExits := []ast.Stmt{stmt}

	// handle any branches; if no label or for me: handle and remove from branches.
	for i := 0; i < len(b.branches); i++ {
		br := b.branches[i]
		if br.Label == nil || br.Label.Obj.Decl.(*ast.LabeledStmt).Stmt == stmt {
			switch br.Tok { // can only be one of these two cases
			case token.CONTINUE:
				b.prev = []ast.Stmt{br}
				b.addSucc(post) // connect to .Post statement if present, for stmt otherwise
			case token.BREAK:
				ctrlExits = append(ctrlExits, br)
			}
			b.branches = append(b.branches[:i], b.branches[i+1:]...)
			i-- // removed in place, so go back to this i
		}
	}

	b.prev = ctrlExits // for stmt and any appropriate break statements
}

// buildSwitch constructs the CFG for switch, type switch, and select statements.
// It handles the initialization (if present), switch expression, and all case clauses.
func (b *builder) buildSwitch(sw ast.Stmt) {
	var cases []ast.Stmt // case 1:, case 2:, ...

	switch sw := sw.(type) {
	case *ast.SwitchStmt: // i.e. switch [ x := 0; ] [ x ] { }
		if sw.Init != nil {
			b.addSucc(sw.Init)
			b.prev = []ast.Stmt{sw.Init}
		}
		b.addSucc(sw)
		b.prev = []ast.Stmt{sw}

		cases = sw.Body.List
	case *ast.TypeSwitchStmt: // i.e. switch [ x := 0; ] t := x.(type) { }
		if sw.Init != nil {
			b.addSucc(sw.Init)
			b.prev = []ast.Stmt{sw.Init}
		}
		b.addSucc(sw)
		b.prev = []ast.Stmt{sw}
		b.addSucc(sw.Assign)
		b.prev = []ast.Stmt{sw.Assign}

		cases = sw.Body.List
	case *ast.SelectStmt: // i.e. select { }
		b.addSucc(sw)
		b.prev = []ast.Stmt{sw}

		cases = sw.Body.List
	}

	var caseExits []ast.Stmt // aggregate of b.prev's resulting from each case
	swPrev := b.prev         // save for each case's previous; Switch or Assign
	var ft *ast.BranchStmt   // fallthrough to handle from previous case, if any
	defaultCase := false

	for _, clause := range cases {
		b.prev = swPrev
		b.addSucc(clause)
		b.prev = []ast.Stmt{clause}
		if ft != nil {
			b.prev = append(b.prev, ft)
		}

		var caseBody []ast.Stmt

		// both of the following cases are guaranteed in spec
		switch clause := clause.(type) {
		case *ast.CaseClause: // i.e. case: [expr,expr,...]:
			if clause.List == nil {
				defaultCase = true
			}
			caseBody = clause.Body
		case *ast.CommClause: // i.e. case c <- chan:
			if clause.Comm == nil {
				defaultCase = true
			} else {
				b.addSucc(clause.Comm)
				b.prev = []ast.Stmt{clause.Comm}
			}
			caseBody = clause.Body
		}

		b.buildBlock(caseBody)

		if ft = fallThrough(caseBody); ft == nil {
			caseExits = append(caseExits, b.prev...)
		}
	}

	if !defaultCase {
		caseExits = append(caseExits, swPrev...)
	}

	// handle any breaks that are unlabeled or for me
	for i := 0; i < len(b.branches); i++ {
		br := b.branches[i]
		if br.Tok == token.BREAK && (br.Label == nil || br.Label.Obj.Decl.(*ast.LabeledStmt).Stmt == sw) {
			caseExits = append(caseExits, br)
			b.branches = append(b.branches[:i], b.branches[i+1:]...)
			i-- // we removed in place, so go back to this index
		}
	}

	b.prev = caseExits // control exits of each case and breaks
}

// fallThrough returns the fallthrough statement at the end of the given slice of statements, if one exists.
// It returns nil if no fallthrough statement is found.
func fallThrough(stmts []ast.Stmt) *ast.BranchStmt {
	if len(stmts) < 1 {
		return nil
	}

	// fallthrough can only be last statement in clause (possibly labeled)
	ft := stmts[len(stmts)-1]

	for { // recursively descend LabeledStmts.
		switch s := ft.(type) {
		case *ast.BranchStmt:
			if s.Tok == token.FALLTHROUGH {
				return s
			}
		case *ast.LabeledStmt:
			ft = s.Stmt
			continue
		}
		break
	}
	return nil
}
