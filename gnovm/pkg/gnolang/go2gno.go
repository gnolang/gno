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
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

func MustReadFile(path string) *FileNode {
	n, err := ReadFile(path)
	if err != nil {
		panic(err)
	}
	return n
}

func MustParseFile(filename string, body string) *FileNode {
	n, err := ParseFile(filename, body)
	if err != nil {
		panic(err)
	}
	return n
}

func ReadFile(path string) (*FileNode, error) {
	bz, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseFile(path, string(bz))
}

func ParseExpr(expr string) (retx Expr, err error) {
	x, err := parser.ParseExpr(expr)
	if err != nil {
		return nil, err
	}
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
	return Go2Gno(nil, x).(Expr), nil
}

func MustParseExpr(expr string) Expr {
	x, err := ParseExpr(expr)
	if err != nil {
		panic(err)
	}
	return x
}

// ParseFile uses the Go parser to parse body. It then runs [Go2Gno] on the
// resulting AST -- the resulting FileNode is returned, together with any other
// error (including panics, which are recovered) from [Go2Gno].
func ParseFile(filename string, body string) (fn *FileNode, err error) {
	// Use go parser to parse the body.
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, filename, body, parser.ParseComments|parser.DeclarationErrors)
	if err != nil {
		return nil, err
	}
	// Print the imports from the file's AST.
	// spew.Dump(f)

	// recover from Go2Gno.
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
	// parse with Go2Gno.
	fn = Go2Gno(fs, f).(*FileNode)
	fn.Name = Name(filename)
	return fn, nil
}

func setLoc(fs *token.FileSet, pos token.Pos, n Node) Node {
	posn := fs.Position(pos)
	n.SetLine(posn.Line)
	return n
}

// If gon is a *ast.File, the name must be filled later.
func Go2Gno(fs *token.FileSet, gon ast.Node) (n Node) {
	if gon == nil {
		return nil
	}
	if fs != nil {
		defer func() {
			if n != nil {
				setLoc(fs, gon.Pos(), n)
			}
		}()
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
		type_ := Go2Gno(fs, gon.Type).(*FuncTypeExpr)
		return &FuncLitExpr{
			Type: *type_,
			Body: toBody(fs, gon.Body),
		}
	case *ast.Field:
		if len(gon.Names) == 0 {
			return &FieldTypeExpr{
				Name: "",
				Type: toExpr(fs, gon.Type),
				Tag:  toExpr(fs, gon.Tag),
			}
		} else if len(gon.Names) == 1 {
			return &FieldTypeExpr{
				Name: toName(gon.Names[0]),
				Type: toExpr(fs, gon.Type),
				Tag:  toExpr(fs, gon.Tag),
			}
		} else {
			panic(fmt.Sprintf(
				"expected a Go Field with 1 name but got %v.\n"+
					"maybe call toFields",
				gon.Names))
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
			Body: toStmts(fs, gon.List),
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
		if cx, ok := gon.X.(*ast.CallExpr); ok {
			if ix, ok := cx.Fun.(*ast.Ident); ok && ix.Name == "panic" {
				if len(cx.Args) != 1 {
					panic("expected panic statement to have single exception value")
				}
				return &PanicStmt{
					Exception: toExpr(fs, cx.Args[0]),
				}
			}
		}
		return &ExprStmt{
			X: toExpr(fs, gon.X),
		}
	case *ast.ForStmt:
		return &ForStmt{
			Init: toSimp(fs, gon.Init),
			Cond: toExpr(fs, gon.Cond),
			Post: toSimp(fs, gon.Post),
			Body: toStmts(fs, gon.Body.List),
		}
	case *ast.IfStmt:
		thenStmt := IfCaseStmt{
			Body: toStmts(fs, gon.Body.List),
		}
		setLoc(fs, gon.Body.Pos(), &thenStmt)
		ess := []Stmt(nil)
		if gon.Else != nil {
			if _, ok := gon.Else.(*ast.BlockStmt); ok {
				ess = Go2Gno(fs, gon.Else).(*BlockStmt).Body
			} else {
				ess = []Stmt{toStmt(fs, gon.Else)}
			}
		}
		elseStmt := IfCaseStmt{
			Body: ess,
		}
		if gon.Else != nil {
			setLoc(fs, gon.Else.Pos(), &elseStmt)
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
			return &SwitchStmt{
				Init:         toStmt(fs, gon.Init),
				X:            toExpr(fs, as.Rhs[0].(*ast.TypeAssertExpr).X),
				IsTypeSwitch: true,
				Clauses:      toClauses(fs, gon.Body.List),
				VarName:      toName(as.Lhs[0].(*ast.Ident)),
			}
		case *ast.ExprStmt:
			return &SwitchStmt{
				Init:         toStmt(fs, gon.Init),
				X:            toExpr(fs, as.X.(*ast.TypeAssertExpr).X),
				IsTypeSwitch: true,
				Clauses:      toClauses(fs, gon.Body.List),
				VarName:      "",
			}
		default:
			panic(fmt.Sprintf(
				"unexpected *ast.TypeSwitchStmt.Assign type %s",
				reflect.TypeOf(gon.Assign).String()))
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
				panic("*ast.FuncDecl cannot have multiple receivers")
			}
			recv = *Go2Gno(fs, gon.Recv.List[0]).(*FieldTypeExpr)
		}
		name := toName(gon.Name)
		type_ := Go2Gno(fs, gon.Type).(*FuncTypeExpr)
		var body []Stmt
		if gon.Body != nil {
			body = Go2Gno(fs, gon.Body).(*BlockStmt).Body
		}
		return &FuncDecl{
			IsMethod: isMethod,
			Recv:     recv,
			NameExpr: NameExpr{Name: name},
			Type:     *type_,
			Body:     body,
		}
	case *ast.GenDecl:
		panic("unexpected *ast.GenDecl; use toDecls(fs,) instead")
	case *ast.File:
		pkgName := Name(gon.Name.Name)
		decls := make([]Decl, 0, len(gon.Decls))
		for _, d := range gon.Decls {
			if gd, ok := d.(*ast.GenDecl); ok {
				decls = append(decls, toDecls(fs, gd)...)
			} else {
				decls = append(decls, toDecl(fs, d))
			}
		}
		return &FileNode{
			Name:    "", // filled later.
			PkgName: pkgName,
			Decls:   decls,
		}
	default:
		panic(fmt.Sprintf("unknown Go type %v: %s\n",
			reflect.TypeOf(gon),
			spew.Sdump(gon),
		))
	}
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
	gnox := Go2Gno(fs, gox)
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
	gnos := Go2Gno(fs, gos)
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
	gnos := Go2Gno(fs, gos)
	if gnos == nil {
		return nil
	} else {
		return gnos.(SimpleStmt).(Stmt)
	}
}

func toDecl(fs *token.FileSet, god ast.Decl) Decl {
	gnod := Go2Gno(fs, god)
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
			ds = append(ds, &TypeDecl{
				NameExpr: NameExpr{Name: name},
				Type:     tipe,
				IsAlias:  alias,
			})
		case *ast.ValueSpec:
			if gd.Tok == token.CONST {
				var names []NameExpr
				var tipe Expr
				var values Exprs
				for _, id := range s.Names {
					names = append(names, *Nx(toName(id)))
				}
				if s.Type == nil {
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
				ds = append(ds, vd)
			}
		case *ast.ImportSpec:
			path, err := strconv.Unquote(s.Path.Value)
			if err != nil {
				panic("unexpected import spec path type")
			}
			ds = append(ds, &ImportDecl{
				NameExpr: *Nx(toName(s.Name)),
				PkgPath:  path,
			})
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
			ftxs = append(ftxs, FieldTypeExpr{
				Name: "",
				Type: toExpr(fs, f.Type),
				Tag:  toExpr(fs, f.Tag),
			})
		} else {
			// one or more named fields
			for _, n := range f.Names {
				ftxs = append(ftxs, FieldTypeExpr{
					Name: toName(n),
					Type: toExpr(fs, f.Type),
					Tag:  toExpr(fs, f.Tag),
				})
			}
		}
	}
	return
}

func toKeyValueExprs(fs *token.FileSet, elts []ast.Expr) (kvxs KeyValueExprs) {
	kvxs = make([]KeyValueExpr, len(elts))
	for i, x := range elts {
		if kvx, ok := x.(*ast.KeyValueExpr); ok {
			kvxs[i] = *Go2Gno(fs, kvx).(*KeyValueExpr)
		} else {
			kvxs[i] = KeyValueExpr{
				Key:   nil,
				Value: toExpr(fs, x),
			}
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
	return SwitchClauseStmt{
		Cases: toExprs(fs, cc.List),
		Body:  toStmts(fs, cc.Body),
	}
}
