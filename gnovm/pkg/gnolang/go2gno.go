package gnolang

/*
	This package helps parse Go code into GNO ast.  It uses go/parser, which
	may imply that the internal gc tooling found in
	`golang/src/cmd/compile/internal/gc|syntax` may not be useful for
	type-checking and static analysis.

	from `golang/src/cmd/compile/README.md`:
	> Note that the `go/*` family of packages, such as `go/parser` and
	> `go/types`, have no relation to the compiler. Since the compiler was
	> initially written in C, the `go/*` packages were developed to enable
	> writing tools working with Go code, such as `gofmt` and `vet`.

	The interpreter is written first and foremost such that we *could* use Go's
	`golang/src/cmd/compile/*` logic for type-checking.  In other words, the
	interpreter is lax and execution may (or may not!) fail if the code run is
	invalid Go code.

	Initially we will use the `go/parser` package to parse Go code and by
	default configure the Machine to perform run-time assertions such as
	type-checking.  Code parsed from `go/parser` that would fail static
	analysis in Go should fail at run-time with the configuration but otherwise
	behave identically.

	This lets us extend the language (e.g. a new kind like the Type kind may
	become available in the Gno language), and helps us plan to transition to
	the final implementation of the Gno parser which should be written in pure
	Gno.  Callers of the interpreter have the option of using the
	`golang/src/cmd/compile/*` package for vetting code correctness.
*/

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/gnolang/gno/gnovm/pkg/parser"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

func (m *Machine) MustReadFile(path string) *FileNode {
	n, err := m.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return n
}

func (m *Machine) MustParseFile(fname string, body string) *FileNode {
	n, err := m.ParseFile(fname, body)
	if err != nil {
		panic(err)
	}
	return n
}

func (m *Machine) ReadFile(path string) (*FileNode, error) {
	bz, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return m.ParseFile(path, string(bz))
}

func (m *Machine) ParseExpr(code string) (expr Expr, err error) {
	x, err := parser.ParseExpr2(code, newParserCallback(m))
	if err != nil {
		return nil, err
	}

	// recover from Go2Gno.
	// NOTE: Go2Gno is best implemented with panics due to inlined toXYZ() calls.
	defer func() {
		if r := recover(); r != nil {
			if rerr, ok := r.(error); ok {
				err = errors.Wrap(rerr, "parsing expression")
			} else {
				err = errors.New(fmt.Sprintf("%v", r)).Stacktrace()
			}
			return
		}
	}()
	// Use a fset, even if empty, so the spans are set properly.
	fset := token.NewFileSet()
	// parse with Go2Gno.
	return Go2Gno(fset, x, nil).(Expr), nil
}

func (m *Machine) MustParseExpr(code string) Expr {
	x, err := m.ParseExpr(code)
	if err != nil {
		panic(err)
	}
	return x
}

func (m *Machine) ParseStmts(code string) (stmts []Stmt, err error) {
	// Go only parses exprs and files,
	// so wrap in a func body.
	fset := token.NewFileSet()
	code = fmt.Sprintf("func(){%s}\n", code)
	x, err := parser.ParseExprFrom2(fset, "<repl>", code, parser.SkipObjectResolution, newParserCallback(m))
	if err != nil {
		return nil, err
	}
	gostmts := x.(*ast.FuncLit).Body.List

	// recover from Go2Gno.
	// NOTE: Go2Gno is best implemented with panics due to inlined toXYZ() calls.
	defer func() {
		if r := recover(); r != nil {
			if rerr, ok := r.(error); ok {
				err = rerr
			} else {
				err = fmt.Errorf("%v", r)
			}
			return
		}
	}()

	// parse with Go2Gno.
	for _, gostmt := range gostmts {
		var stmt Stmt
		nn := Go2Gno(fset, gostmt, nil)
		switch nn := nn.(type) {
		case Stmt:
			stmt = nn
		case Expr:
			stmt = &ExprStmt{X: nn}
		default:
			panic(fmt.Sprintf(
				"unexpected AST type %v (%T)", nn, nn))
		}
		stmts = append(stmts, stmt)
	}
	return stmts, nil
}

func (m *Machine) MustParseStmts(code string) []Stmt {
	stmts, err := m.ParseStmts(code)
	if err != nil {
		panic(err)
	}
	return stmts
}

func (m *Machine) ParseDecls(code string) (decls []Decl, err error) {
	// Go only parses exprs and files,
	// so wrap in a func body.
	code = fmt.Sprintf("package repl\n%s\n", code)
	fn, err := m.ParseFile("<repl>", code)
	if err != nil {
		return nil, err
	}
	return fn.Decls, nil
}

func (m *Machine) MustParseDecls(code string) []Decl {
	decls, err := m.ParseDecls(code)
	if err != nil {
		panic(err)
	}
	return decls
}

// ParseFilePackageName returns the package name of a gno file.
func ParseFilePackageName(fname string) (string, error) {
	fs := token.NewFileSet()
	// Just parse the package clause, and nothing else.
	f, err := parser.ParseFile(fs, fname, nil, parser.PackageClauseOnly)
	if err != nil {
		return "", err
	}
	return f.Name.Name, nil
}

const (
	tokenCostFactor   = 1 // To be adjusted from benchmarks.
	nestingCostFactor = 1 // To be adjusted from benchmarks.
)

func newParserCallback(m *Machine) parser.ParserCallback {
	if m == nil || m.GasMeter == nil {
		return nil
	}
	return func(tok token.Token, nestLev int) {
		m.GasMeter.ConsumeGas(types.Gas(tokenCostFactor+nestLev*nestingCostFactor), "parsing")
	}
}

// ParseFile uses the Go parser to parse body. It then runs [Go2Gno] on the
// resulting AST -- the resulting FileNode is returned, together with any other
// error (including panics, which are recovered) from [Go2Gno].
func (m *Machine) ParseFile(fname string, body string) (fn *FileNode, err error) {
	// Use go parser to parse the body.
	fs := token.NewFileSet()
	// TODO(morgan): would be nice to add parser.SkipObjectResolution as we don't
	// seem to be using its features, but this breaks when testing (specifically redeclaration tests).
	const parseOpts = parser.ParseComments | parser.DeclarationErrors
	astf, err := parser.ParseFile2(fs, fname, body, parseOpts, newParserCallback(m))
	if err != nil {
		return nil, err
	}
	// Print the imports from the file's AST.
	// spew.Dump(f)

	// NOTE: DO NOT Disable this when running with -debug or similar.
	// Global environment variables are a vector for attack and should not
	// be relied upon for critical production systems.  Recovers from
	// Go2Gno and returns an error.
	// NOTE: Go2Gno is best implemented with panics due to inlined toXYZ() calls.
	defer func() {
		if r := recover(); r != nil {
			if rerr, ok := r.(error); ok {
				err = errors.Wrap(rerr, "parsing file")
			} else {
				err = errors.New(fmt.Sprintf("%v", r)).Stacktrace()
			}
			return
		}
	}()
	// Parse with Go2Gno.
	fn = Go2Gno(fs, astf, nil).(*FileNode)
	fn.FileName = fname
	return fn, nil
}

// setSpan() will not attempt to overwrite an existing span.
// This usually happens when an inner node is passed outward,
// in which case we want to keep the original specificity.
func setSpan(fs *token.FileSet, gon ast.Node, n Node) Node {
	if n.GetSpan().IsZero() {
		n.SetSpan(SpanFromGo(fs, gon))
	}
	return n
}

// If gon is a *ast.File, the name must be filled later.
func Go2Gno(fs *token.FileSet, gon ast.Node, fileComments []*ast.CommentGroup) (n Node) {
	if gon == nil {
		return nil
	}
	if fs != nil {
		defer func() {
			if n != nil {
				setSpan(fs, gon, n)
			}
		}()
	}

	panicWithPos := func(fmtStr string, args ...any) {
		pos := fs.Position(gon.Pos())
		loc := fmt.Sprintf("%s:%d:%d", pos.Filename, pos.Line, pos.Column)
		panic(fmt.Errorf("%s: %v", loc, fmt.Sprintf(fmtStr, args...)))
	}

	switch gon := gon.(type) {
	case *ast.ParenExpr:
		return toExpr(fs, gon.X)
	case *ast.Ident:
		return Nx(toName(gon))
	case *ast.BasicLit:
		if gon == nil {
			return nil
		}
		return &BasicLitExpr{
			Kind:  toWord(gon.Kind),
			Value: gon.Value,
		}
	case *ast.BinaryExpr:
		return &BinaryExpr{
			Left:  toExpr(fs, gon.X),
			Op:    toWord(gon.Op),
			Right: toExpr(fs, gon.Y),
		}
	case *ast.CallExpr:
		return &CallExpr{
			Func: toExpr(fs, gon.Fun),
			Args: toExprs(fs, gon.Args),
			Varg: gon.Ellipsis.IsValid(),
		}
	case *ast.IndexExpr:
		return &IndexExpr{
			X:     toExpr(fs, gon.X),
			Index: toExpr(fs, gon.Index),
		}
	case *ast.SelectorExpr:
		return &SelectorExpr{
			X:   toExpr(fs, gon.X),
			Sel: toName(gon.Sel),
		}
	case *ast.SliceExpr:
		return &SliceExpr{
			X:    toExpr(fs, gon.X),
			Low:  toExpr(fs, gon.Low),
			High: toExpr(fs, gon.High),
			Max:  toExpr(fs, gon.Max),
		}
	case *ast.StarExpr:
		return &StarExpr{
			X: toExpr(fs, gon.X),
		}
	case *ast.TypeAssertExpr:
		return &TypeAssertExpr{
			X:    toExpr(fs, gon.X),
			Type: toExpr(fs, gon.Type),
		}
	case *ast.UnaryExpr:
		if gon.Op == token.AND {
			return &RefExpr{
				X: toExpr(fs, gon.X),
			}
		} else {
			return &UnaryExpr{
				X:  toExpr(fs, gon.X),
				Op: toWord(gon.Op),
			}
		}
	case *ast.CompositeLit:
		// If ArrayType with ellipsis for length,
		// just figure out the length here.
		return &CompositeLitExpr{
			Type: toExpr(fs, gon.Type),
			Elts: toKeyValueExprs(fs, gon.Elts),
		}
	case *ast.KeyValueExpr:
		return &KeyValueExpr{
			Key:   toExpr(fs, gon.Key),
			Value: toExpr(fs, gon.Value),
		}
	case *ast.FuncLit:
		type_ := Go2Gno(fs, gon.Type, fileComments).(*FuncTypeExpr)

		return &FuncLitExpr{
			Type: *type_,
			Body: toBody(fs, gon.Body),
		}
	case *ast.Field:
		if len(gon.Names) == 0 {
			return &FieldTypeExpr{
				NameExpr: *Nx(""),
				Type:     toExpr(fs, gon.Type),
				Tag:      toExpr(fs, gon.Tag),
			}
		} else if len(gon.Names) == 1 {
			return &FieldTypeExpr{
				NameExpr: *Nx(toName(gon.Names[0])),
				Type:     toExpr(fs, gon.Type),
				Tag:      toExpr(fs, gon.Tag),
			}
		} else {
			panicWithPos(
				"expected a Go Field with 1 name but got %v.\n"+
					"maybe call toFields",
				gon.Names)
		}
	case *ast.ArrayType:
		if _, ok := gon.Len.(*ast.Ellipsis); ok {
			return &ArrayTypeExpr{
				Len: nil,
				Elt: toExpr(fs, gon.Elt),
			}
		} else if gon.Len == nil {
			return &SliceTypeExpr{
				Elt: toExpr(fs, gon.Elt),
				Vrd: false,
			}
		} else {
			return &ArrayTypeExpr{
				Len: toExpr(fs, gon.Len),
				Elt: toExpr(fs, gon.Elt),
			}
		}
	case *ast.Ellipsis:
		return &SliceTypeExpr{
			Elt: toExpr(fs, gon.Elt),
			Vrd: true,
		}
	case *ast.InterfaceType:
		return &InterfaceTypeExpr{
			Methods: toFieldsFromList(fs, gon.Methods),
		}
	case *ast.ChanType:
		var dir ChanDir
		if gon.Dir&ast.SEND > 0 {
			dir |= SEND
		}
		if gon.Dir&ast.RECV > 0 {
			dir |= RECV
		}
		return &ChanTypeExpr{
			Dir:   dir,
			Value: toExpr(fs, gon.Value),
		}
	case *ast.FuncType:
		return &FuncTypeExpr{
			Params:  toFieldsFromList(fs, gon.Params),
			Results: toFieldsFromList(fs, gon.Results),
		}
	case *ast.MapType:
		return &MapTypeExpr{
			Key:   toExpr(fs, gon.Key),
			Value: toExpr(fs, gon.Value),
		}
	case *ast.StructType:
		return &StructTypeExpr{
			Fields: toFieldsFromList(fs, gon.Fields),
		}
	case *ast.AssignStmt:
		return &AssignStmt{
			Lhs: toExprs(fs, gon.Lhs),
			Op:  toWord(gon.Tok),
			Rhs: toExprs(fs, gon.Rhs),
		}
	case *ast.BlockStmt:
		return &BlockStmt{
			Body: toBody(fs, gon),
		}
	case *ast.BranchStmt:
		return &BranchStmt{
			Op:    toWord(gon.Tok),
			Label: toName(gon.Label),
		}
	case *ast.DeclStmt:
		return &DeclStmt{
			Body: toSimpleDeclStmts(fs, gon.Decl.(*ast.GenDecl)),
		}
	case *ast.DeferStmt:
		cx := toExpr(fs, gon.Call).(*CallExpr)
		return &DeferStmt{
			Call: *cx,
		}
	case *ast.ExprStmt:
		return &ExprStmt{
			X: toExpr(fs, gon.X),
		}
	case *ast.ForStmt:
		return &ForStmt{
			Init: toSimp(fs, gon.Init),
			Cond: toExpr(fs, gon.Cond),
			Post: toSimp(fs, gon.Post),
			Body: toBody(fs, gon.Body),
		}
	case *ast.IfStmt:
		thenStmt := IfCaseStmt{
			Body: toBody(fs, gon.Body),
		}
		setSpan(fs, gon.Body, &thenStmt)
		ess := []Stmt(nil)
		if gon.Else != nil {
			if _, ok := gon.Else.(*ast.BlockStmt); ok {
				ess = Go2Gno(fs, gon.Else, fileComments).(*BlockStmt).Body
			} else {
				ess = []Stmt{toStmt(fs, gon.Else)}
			}
		}
		elseStmt := IfCaseStmt{
			Body: ess,
		}
		if gon.Else != nil {
			setSpan(fs, gon.Else, &elseStmt)
		}
		return &IfStmt{
			Init: toSimp(fs, gon.Init),
			Cond: toExpr(fs, gon.Cond),
			Then: thenStmt,
			Else: elseStmt,
		}
	case *ast.IncDecStmt:
		return &IncDecStmt{
			X:  toExpr(fs, gon.X),
			Op: toWord(gon.Tok),
		}
	case *ast.LabeledStmt:
		stmt := toStmt(fs, gon.Stmt)
		stmt.SetLabel(toName(gon.Label))
		return stmt
	case *ast.RangeStmt:
		return &RangeStmt{
			X:     toExpr(fs, gon.X),
			Key:   toExpr(fs, gon.Key),
			Value: toExpr(fs, gon.Value),
			Op:    toWord(gon.Tok),
			Body:  toBody(fs, gon.Body),
		}
	case *ast.ReturnStmt:
		return &ReturnStmt{
			Results: toExprs(fs, gon.Results),
		}
	case *ast.TypeSwitchStmt:
		switch as := gon.Assign.(type) {
		case *ast.AssignStmt:
			stmt := &SwitchStmt{
				Init:         toStmt(fs, gon.Init),
				X:            toExpr(fs, as.Rhs[0].(*ast.TypeAssertExpr).X),
				IsTypeSwitch: true,
				Clauses:      toClauses(fs, gon.Body.List),
				VarName:      toName(as.Lhs[0].(*ast.Ident)),
			}
			return stmt
		case *ast.ExprStmt:
			stmt := &SwitchStmt{
				Init:         toStmt(fs, gon.Init),
				X:            toExpr(fs, as.X.(*ast.TypeAssertExpr).X),
				IsTypeSwitch: true,
				Clauses:      toClauses(fs, gon.Body.List),
				VarName:      "",
			}
			return stmt
		default:
			panicWithPos("unexpected *ast.TypeSwitchStmt.Assign type %s",
				reflect.TypeOf(gon.Assign).String())
		}
	case *ast.SwitchStmt:
		x := toExpr(fs, gon.Tag)
		if x == nil {
			// if tag is nil, default to "true"
			x = Nx(Name("true"))
		}
		return &SwitchStmt{
			Init:         toStmt(fs, gon.Init),
			X:            x,
			IsTypeSwitch: false,
			Clauses:      toClauses(fs, gon.Body.List),
		}
	case *ast.FuncDecl:
		isMethod := gon.Recv != nil
		recv := FieldTypeExpr{}
		if isMethod {
			if len(gon.Recv.List) > 1 {
				panicWithPos("method has multiple receivers")
			}
			if len(gon.Recv.List) == 0 {
				panicWithPos("method has no receiver")
			}
			recv = *Go2Gno(fs, gon.Recv.List[0], fileComments).(*FieldTypeExpr)
		}
		name := toName(gon.Name)
		type_ := Go2Gno(fs, gon.Type, fileComments).(*FuncTypeExpr)
		var body []Stmt
		if gon.Body != nil {
			body = Go2Gno(fs, gon.Body, fileComments).(*BlockStmt).Body
		}
		fd := &FuncDecl{
			IsMethod: isMethod,
			Recv:     recv,
			NameExpr: NameExpr{Name: name},
			Type:     *type_,
			Body:     body,
		}
		if gon.Body != nil && strings.HasPrefix(gon.Name.Name, "Example") && fileComments != nil {
			output, unordered, hasOutput := exampleOutput(gon.Body, fileComments)
			if hasOutput {
				fd.SetAttribute(ATTR_EXAMPLE_OUTPUT, output)
				fd.SetAttribute(ATTR_OUTPUT_UNORDERED, unordered)
			}
		}
		return fd
	case *ast.GenDecl:
		panicWithPos("unexpected *ast.GenDecl; use toDecls(fs,) instead")
	case *ast.File:
		pkgName := Name(gon.Name.Name)
		decls := make([]Decl, 0, len(gon.Decls))
		for _, d := range gon.Decls {
			if gd, ok := d.(*ast.GenDecl); ok {
				decls = append(decls, toDecls(fs, gd)...)
			} else {
				decls = append(decls, toDecl(fs, d, gon.Comments))
			}
		}
		return &FileNode{
			FileName: "", // filled later.
			PkgName:  pkgName,
			Decls:    decls,
		}
	case *ast.EmptyStmt:
		return &EmptyStmt{}
	case *ast.IndexListExpr:
		if len(gon.Indices) > 1 {
			panicWithPos("invalid operation: more than one index")
		}
		panicWithPos("invalid operation: indexList is not permitted in Gno")
	case *ast.GoStmt:
		panicWithPos("goroutines are not permitted")
	default:
		panicWithPos("unknown Go type %v: %s\n",
			reflect.TypeOf(gon),
			spew.Sdump(gon),
		)
	}

	return
}

//----------------------------------------
// utility methods

func toName(name *ast.Ident) Name {
	if name == nil {
		return Name("")
	} else {
		return Name(name.Name)
	}
}

var token2word = map[token.Token]Word{
	token.ILLEGAL:        ILLEGAL,
	token.IDENT:          NAME,
	token.INT:            INT,
	token.FLOAT:          FLOAT,
	token.IMAG:           IMAG,
	token.CHAR:           CHAR,
	token.STRING:         STRING,
	token.ADD:            ADD,
	token.SUB:            SUB,
	token.MUL:            MUL,
	token.QUO:            QUO,
	token.REM:            REM,
	token.AND:            BAND,
	token.OR:             BOR,
	token.XOR:            XOR,
	token.SHL:            SHL,
	token.SHR:            SHR,
	token.AND_NOT:        BAND_NOT,
	token.ADD_ASSIGN:     ADD_ASSIGN,
	token.SUB_ASSIGN:     SUB_ASSIGN,
	token.MUL_ASSIGN:     MUL_ASSIGN,
	token.QUO_ASSIGN:     QUO_ASSIGN,
	token.REM_ASSIGN:     REM_ASSIGN,
	token.AND_ASSIGN:     BAND_ASSIGN,
	token.OR_ASSIGN:      BOR_ASSIGN,
	token.XOR_ASSIGN:     XOR_ASSIGN,
	token.SHL_ASSIGN:     SHL_ASSIGN,
	token.SHR_ASSIGN:     SHR_ASSIGN,
	token.AND_NOT_ASSIGN: BAND_NOT_ASSIGN,
	token.LAND:           LAND,
	token.LOR:            LOR,
	token.ARROW:          ARROW,
	token.INC:            INC,
	token.DEC:            DEC,
	token.EQL:            EQL,
	token.LSS:            LSS,
	token.GTR:            GTR,
	token.ASSIGN:         ASSIGN,
	token.NOT:            NOT,
	token.NEQ:            NEQ,
	token.LEQ:            LEQ,
	token.GEQ:            GEQ,
	token.DEFINE:         DEFINE,
	token.BREAK:          BREAK,
	token.CASE:           CASE,
	token.CHAN:           CHAN,
	token.CONST:          CONST,
	token.CONTINUE:       CONTINUE,
	token.DEFAULT:        DEFAULT,
	token.DEFER:          DEFER,
	token.ELSE:           ELSE,
	token.FALLTHROUGH:    FALLTHROUGH,
	token.FOR:            FOR,
	token.FUNC:           FUNC,
	token.GO:             GO,
	token.GOTO:           GOTO,
	token.IF:             IF,
	token.IMPORT:         IMPORT,
	token.RETURN:         RETURN,
	token.SELECT:         SELECT,
	token.STRUCT:         STRUCT,
	token.TYPE:           TYPE,
	token.VAR:            VAR,
}

func toWord(tok token.Token) Word {
	return token2word[tok]
}

func toExpr(fs *token.FileSet, gox ast.Expr) Expr {
	// TODO: could the language handle this?
	gnox := Go2Gno(fs, gox, nil)
	if gnox == nil {
		return nil
	} else {
		return gnox.(Expr)
	}
}

func toExprs(fs *token.FileSet, goxs []ast.Expr) (gnoxs Exprs) {
	if len(goxs) == 0 {
		return nil
	}
	gnoxs = make([]Expr, len(goxs))
	for i, x := range goxs {
		gnoxs[i] = toExpr(fs, x)
	}
	return
}

func toStmt(fs *token.FileSet, gos ast.Stmt) Stmt {
	gnos := Go2Gno(fs, gos, nil)
	if gnos == nil {
		return nil
	} else {
		return gnos.(Stmt)
	}
}

func toStmts(fs *token.FileSet, goss []ast.Stmt) (gnoss Body) {
	gnoss = make([]Stmt, len(goss))
	for i, x := range goss {
		gnoss[i] = toStmt(fs, x)
	}
	return
}

func toBody(fs *token.FileSet, body *ast.BlockStmt) Body {
	if body == nil {
		return nil
	}
	return toStmts(fs, body.List)
}

func toSimp(fs *token.FileSet, gos ast.Stmt) Stmt {
	gnos := Go2Gno(fs, gos, nil)
	if gnos == nil {
		return nil
	} else {
		return gnos.(SimpleStmt).(Stmt)
	}
}

func toDecl(fs *token.FileSet, god ast.Decl, fileComments []*ast.CommentGroup) Decl {
	gnod := Go2Gno(fs, god, fileComments)
	if gnod == nil {
		return nil
	} else {
		return gnod.(Decl)
	}
}

func toDecls(fs *token.FileSet, gd *ast.GenDecl) (ds Decls) {
	ds = make([]Decl, 0, len(gd.Specs))
	/*
		Within a parenthesized const declaration list the
		expression list may be omitted from any but the
		first ConstSpec. Such an empty list is equivalent
		to the textual substitution of the first preceding
		non-empty expression list and its type if any.
	*/
	var lastValues Exprs // (see Go iota spec above)
	var lastType Expr    // (see Go iota spec above)
	for si, s := range gd.Specs {
		switch s := s.(type) {
		case *ast.TypeSpec:
			name := toName(s.Name)
			tipe := toExpr(fs, s.Type)
			alias := s.Assign != 0
			td := &TypeDecl{
				NameExpr: NameExpr{Name: name},
				Type:     tipe,
				IsAlias:  alias,
			}
			setSpan(fs, s, td)
			ds = append(ds, td)
		case *ast.ValueSpec:
			if gd.Tok == token.CONST {
				var names []NameExpr
				var tipe Expr
				var values Exprs
				for _, id := range s.Names {
					names = append(names, *Nx(toName(id)))
				}

				// Inherit the last type when
				// both type and value are nil
				if s.Type == nil && s.Values == nil {
					tipe = lastType
				} else {
					tipe = toExpr(fs, s.Type)
					lastType = tipe
				}

				if s.Values == nil {
					values = copyExprs(lastValues)
				} else {
					values = toExprs(fs, s.Values)
					lastValues = values
				}
				cd := &ValueDecl{
					NameExprs: names,
					Type:      tipe,
					Values:    values,
					Const:     true,
				}
				cd.SetAttribute(ATTR_IOTA, si)
				setSpan(fs, s, cd)
				ds = append(ds, cd)
			} else {
				var names []NameExpr
				var tipe Expr
				var values Exprs
				for _, id := range s.Names {
					names = append(names, *Nx(toName(id)))
				}
				tipe = toExpr(fs, s.Type)
				if s.Values != nil {
					values = toExprs(fs, s.Values)
				}
				vd := &ValueDecl{
					NameExprs: names,
					Type:      tipe,
					Values:    values,
					Const:     false,
				}
				setSpan(fs, s, vd)
				ds = append(ds, vd)
			}
		case *ast.ImportSpec:
			path, err := strconv.Unquote(s.Path.Value)
			if err != nil {
				panic("unexpected import spec path type")
			}
			im := &ImportDecl{
				NameExpr: *Nx(toName(s.Name)),
				PkgPath:  path,
			}
			setSpan(fs, s, im)
			ds = append(ds, im)
		default:
			panic(fmt.Sprintf(
				"unexpected decl spec %v",
				reflect.TypeOf(s)))
		}
	}

	return ds
}

func toSimpleDeclStmts(fs *token.FileSet, gd *ast.GenDecl) (sds []Stmt) {
	ds := toDecls(fs, gd)
	sds = make([]Stmt, len(ds))
	for i, d := range ds {
		sds[i] = d.(SimpleDeclStmt).(Stmt)
	}
	return
}

func toFieldsFromList(fs *token.FileSet, fl *ast.FieldList) (ftxs []FieldTypeExpr) {
	if fl == nil {
		return nil
	} else {
		ftxs = toFields(fs, fl.List...)
		return
	}
}

func toFields(fs *token.FileSet, fields ...*ast.Field) (ftxs []FieldTypeExpr) {
	if len(fields) == 0 {
		return nil
	}
	ftxs = make([]FieldTypeExpr, 0, len(fields)) // may grow longer
	for _, f := range fields {
		if len(f.Names) == 0 {
			// a single unnamed field w/ type
			ftx := FieldTypeExpr{
				NameExpr: *Nx(""),
				Type:     toExpr(fs, f.Type),
				Tag:      toExpr(fs, f.Tag),
			}
			setSpan(fs, f, &ftx)
			ftxs = append(ftxs, ftx)
		} else {
			// one or more named fields
			for _, n := range f.Names {
				ftx := FieldTypeExpr{
					NameExpr: *Nx(toName(n)),
					Type:     toExpr(fs, f.Type),
					Tag:      toExpr(fs, f.Tag),
				}
				setSpan(fs, f, &ftx)
				ftxs = append(ftxs, ftx)
			}
		}
	}
	return
}

func toKeyValueExprs(fs *token.FileSet, elts []ast.Expr) (kvxs KeyValueExprs) {
	kvxs = make([]KeyValueExpr, len(elts))
	for i, x := range elts {
		if kvx, ok := x.(*ast.KeyValueExpr); ok {
			kvxs[i] = *Go2Gno(fs, kvx, nil).(*KeyValueExpr)
		} else {
			kvx := KeyValueExpr{
				Key:   nil,
				Value: toExpr(fs, x),
			}
			setSpan(fs, x, &kvx)
			kvxs[i] = kvx
		}
	}
	return
}

// NOTE: moves the default clause to last.
func toClauses(fs *token.FileSet, csz []ast.Stmt) []SwitchClauseStmt {
	res := make([]SwitchClauseStmt, 0, len(csz))
	var dclause *SwitchClauseStmt
	for _, cs := range csz {
		clause := toSwitchClauseStmt(fs, cs.(*ast.CaseClause))
		if len(clause.Cases) == 0 {
			if dclause != nil {
				panic("duplicate default clause")
			}
			dclause = &clause
		} else {
			res = append(res, clause)
		}
	}
	if dclause != nil {
		res = append(res, *dclause)
	}
	if len(res) != len(csz) {
		panic("should not happen")
	}
	return res
}

func toSwitchClauseStmt(fs *token.FileSet, cc *ast.CaseClause) SwitchClauseStmt {
	scs := SwitchClauseStmt{
		Cases: toExprs(fs, cc.List),
		Body:  toStmts(fs, cc.Body),
	}
	setSpan(fs, cc, &scs)
	return scs
}

// From https://github.com/golang/go/blob/master/src/cmd/go/internal/load/test.go

var outputPrefix = regexp.MustCompile(`(?i)^[[:space:]]*(unordered )?output:`)

// exampleOutput extracts the expected output and whether there was a valid output comment.
func exampleOutput(b *ast.BlockStmt, comments []*ast.CommentGroup) (output string, unordered, ok bool) {
	if _, last := lastComment(b, comments); last != nil {
		// test that it begins with the correct prefix
		text := last.Text()
		if loc := outputPrefix.FindStringSubmatchIndex(text); loc != nil {
			if loc[2] != -1 {
				unordered = true
			}
			text = text[loc[1]:]
			// Strip zero or more spaces followed by \n or a single space.
			text = strings.TrimLeft(text, " ")
			if len(text) > 0 && text[0] == '\n' {
				text = text[1:]
			}
			return text, unordered, true
		}
	}
	return "", false, false // no suitable comment found
}

// lastComment returns the last comment inside the provided block.
func lastComment(b *ast.BlockStmt, c []*ast.CommentGroup) (i int, last *ast.CommentGroup) {
	if b == nil {
		return
	}
	pos, end := b.Pos(), b.End()
	for j, cg := range c {
		if cg.Pos() < pos {
			continue
		}
		if cg.End() > end {
			break
		}
		i, last = j, cg
	}
	return
}
