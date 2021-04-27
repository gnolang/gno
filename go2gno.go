package gno

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

	This lets us extend the langauge (e.g. a new kind like the Type kind may
	become available in the Gno langauge), and helps us plan to transition to
	the final implementation of the Gno parser which should be written in pure
	Gno.  Callers of the interpreter have the option of using the
	`golang/src/cmd/compile/*` package for vetting code correctness.
*/

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"reflect"
	"strconv"

	"github.com/davecgh/go-spew/spew"
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
	bz, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseFile(path, string(bz))
}

// filename must not include the path.
func ParseFile(filename string, body string) (*FileNode, error) {

	// Parse src but stop after processing the imports.
	f, err := parser.ParseFile(token.NewFileSet(), filename, body, parser.ParseComments|parser.DeclarationErrors)
	if err != nil {
		return nil, err
	}

	// Print the imports from the file's AST.
	// spew.Dump(f)
	fn := Go2Gno(f).(*FileNode)
	fn.Name = Name(filename)
	return fn, nil
}

// If gon is a *ast.File, the name must be filled later.
func Go2Gno(gon ast.Node) (n Node) {
	if gon == nil {
		return nil
	}
	switch gon := gon.(type) {
	case *ast.File:
		pkgName := Name(gon.Name.Name)
		decls := make([]Decl, 0, len(gon.Decls))
		for _, d := range gon.Decls {
			if gd, ok := d.(*ast.GenDecl); ok {
				decls = append(decls, toDecls(gd)...)
			} else {
				decls = append(decls, toDecl(d))
			}
		}
		return &FileNode{
			Name:    "", // filled later.
			PkgName: pkgName,
			Decls:   decls,
		}
	case *ast.FuncDecl:
		isMethod := gon.Recv != nil
		recv := FieldTypeExpr{}
		if isMethod {
			if len(gon.Recv.List) > 1 {
				panic("*ast.FuncDecl cannot have multiple receivers")
			}
			recv = *Go2Gno(gon.Recv.List[0]).(*FieldTypeExpr)
		}
		name := toName(gon.Name)
		type_ := Go2Gno(gon.Type).(*FuncTypeExpr)
		body := Go2Gno(gon.Body).(*BlockStmt).Body
		return &FuncDecl{
			IsMethod: isMethod,
			Recv:     recv,
			NameExpr: NameExpr{Name: name},
			Type:     *type_,
			Body:     body,
		}
	case *ast.FuncType:
		return &FuncTypeExpr{
			Params:  toFieldsFromList(gon.Params),
			Results: toFieldsFromList(gon.Results),
		}
	case *ast.BlockStmt:
		return &BlockStmt{
			Body: toStmts(gon.List),
		}
	case *ast.ForStmt:
		return &ForStmt{
			Init: toSimp(gon.Init),
			Cond: toExpr(gon.Cond),
			Post: toSimp(gon.Post),
			Body: toStmts(gon.Body.List),
		}
	case *ast.AssignStmt:
		return &AssignStmt{
			Lhs: toExprs(gon.Lhs),
			Op:  toWord(gon.Tok),
			Rhs: toExprs(gon.Rhs),
		}
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
			Left:  toExpr(gon.X),
			Op:    toWord(gon.Op),
			Right: toExpr(gon.Y),
		}
	case *ast.IncDecStmt:
		return &IncDecStmt{
			X:  toExpr(gon.X),
			Op: toWord(gon.Tok),
		}
	case *ast.IfStmt:
		ess := []Stmt(nil)
		if gon.Else != nil {
			if _, ok := gon.Else.(*ast.BlockStmt); ok {
				ess = Go2Gno(gon.Else).(*BlockStmt).Body
			} else {
				ess = []Stmt{toStmt(gon.Else)}
			}
		}
		return &IfStmt{
			Init: toSimp(gon.Init),
			Cond: toExpr(gon.Cond),
			Then: IfCaseStmt{
				Body: toStmts(gon.Body.List),
			},
			Else: IfCaseStmt{
				Body: ess,
			},
		}
	case *ast.UnaryExpr:
		if gon.Op == token.AND {
			return &RefExpr{
				X: toExpr(gon.X),
			}
		} else {
			return &UnaryExpr{
				X:  toExpr(gon.X),
				Op: toWord(gon.Op),
			}
		}
	case *ast.ReturnStmt:
		return &ReturnStmt{
			Results: toExprs(gon.Results),
		}
	case *ast.Field:
		if len(gon.Names) == 0 {
			return &FieldTypeExpr{
				Name: "",
				Type: toExpr(gon.Type),
				Tag:  toExpr(gon.Tag),
			}
		} else if len(gon.Names) == 1 {
			return &FieldTypeExpr{
				Name: toName(gon.Names[0]),
				Type: toExpr(gon.Type),
				Tag:  toExpr(gon.Tag),
			}
		} else {
			panic(fmt.Sprintf(
				"expected a Go Field with 1 name but got %v.\n"+
					"maybe call toFields",
				gon.Names))
		}
	case *ast.StructType:
		return &StructTypeExpr{
			Fields: toFieldsFromList(gon.Fields),
		}
	case *ast.InterfaceType:
		return &InterfaceTypeExpr{
			Methods: toFieldsFromList(gon.Methods),
		}
	case *ast.GenDecl:
		panic("unexpected *ast.GenDecl; use toDecls() instead")
	case *ast.CompositeLit:
		// If ArrayType with ellipsis for length,
		// just figure out the length here.
		return &CompositeLitExpr{
			Type: toExpr(gon.Type),
			Elts: toKeyValueExprs(gon.Elts),
		}
	case *ast.ExprStmt:
		return &ExprStmt{
			X: toExpr(gon.X),
		}
	case *ast.CallExpr:
		return &CallExpr{
			Func: toExpr(gon.Fun),
			Args: toExprs(gon.Args),
			Varg: gon.Ellipsis.IsValid(),
		}
	case *ast.KeyValueExpr:
		return &KeyValueExpr{
			Key:   toExpr(gon.Key),
			Value: toExpr(gon.Value),
		}
	case *ast.SelectorExpr:
		return &SelectorExpr{
			X:   toExpr(gon.X),
			Sel: toName(gon.Sel),
		}
	case *ast.Ellipsis:
		return &SliceTypeExpr{
			Elt: toExpr(gon.Elt),
			Vrd: true,
		}
	case *ast.ArrayType:
		if _, ok := gon.Len.(*ast.Ellipsis); ok {
			return &ArrayTypeExpr{
				Len: nil,
				Elt: toExpr(gon.Elt),
			}
		} else if gon.Len == nil {
			return &SliceTypeExpr{
				Elt: toExpr(gon.Elt),
				Vrd: false,
			}
		} else {
			return &ArrayTypeExpr{
				Len: toExpr(gon.Len),
				Elt: toExpr(gon.Elt),
			}
		}
	case *ast.IndexExpr:
		return &IndexExpr{
			X:     toExpr(gon.X),
			Index: toExpr(gon.Index),
		}
	case *ast.RangeStmt:
		return &RangeStmt{
			X:     toExpr(gon.X),
			Key:   toExpr(gon.Key),
			Value: toExpr(gon.Value),
			Op:    toWord(gon.Tok),
			Body:  toBody(gon.Body),
		}
	case *ast.BranchStmt:
		return &BranchStmt{
			Op:    toWord(gon.Tok),
			Label: toName(gon.Label),
		}
	case *ast.DeclStmt:
		return &DeclStmt{
			Decls: toSimpleDecls(gon.Decl.(*ast.GenDecl)),
		}
	case *ast.SliceExpr:
		return &SliceExpr{
			X:    toExpr(gon.X),
			Low:  toExpr(gon.Low),
			High: toExpr(gon.High),
			Max:  toExpr(gon.Max),
		}
	case *ast.StarExpr:
		return &StarExpr{
			X: toExpr(gon.X),
		}
	case *ast.TypeAssertExpr:
		return &TypeAssertExpr{
			X:    toExpr(gon.X),
			Type: toExpr(gon.Type),
		}
	case *ast.ParenExpr:
		return toExpr(gon.X)
	case *ast.MapType:
		return &MapTypeExpr{
			Key:   toExpr(gon.Key),
			Value: toExpr(gon.Value),
		}
	case *ast.FuncLit:
		type_ := Go2Gno(gon.Type).(*FuncTypeExpr)
		return &FuncLitExpr{
			Type: *type_,
			Body: toBody(gon.Body),
		}
	case *ast.DeferStmt:
		cx := toExpr(gon.Call).(*CallExpr)
		return &DeferStmt{
			Call: *cx,
		}
	case *ast.LabeledStmt:
		return &LabeledStmt{
			Label: toName(gon.Label),
			Stmt:  toStmt(gon.Stmt),
		}
	case *ast.TypeSwitchStmt:
		switch as := gon.Assign.(type) {
		case *ast.AssignStmt:
			return &SwitchStmt{
				Init:         toStmt(gon.Init),
				X:            toExpr(as.Rhs[0].(*ast.TypeAssertExpr).X),
				IsTypeSwitch: true,
				Clauses:      toClauses(gon.Body.List),
				VarName:      toName(as.Lhs[0].(*ast.Ident)),
			}
		case *ast.ExprStmt:
			return &SwitchStmt{
				Init:         toStmt(gon.Init),
				X:            toExpr(as.X.(*ast.TypeAssertExpr).X),
				IsTypeSwitch: true,
				Clauses:      toClauses(gon.Body.List),
				VarName:      "",
			}
		default:
			panic(fmt.Sprintf(
				"unexpected *ast.TypeSwitchStmt.Assign type %s",
				reflect.TypeOf(gon.Assign).String()))
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
			Value: toExpr(gon.Value),
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

func toExpr(gox ast.Expr) Expr {
	// TODO: could the language handle this?
	gnox := Go2Gno(gox)
	if gnox == nil {
		return nil
	} else {
		return gnox.(Expr)
	}
}

func toExprs(goxs []ast.Expr) (gnoxs Exprs) {
	gnoxs = make([]Expr, len(goxs))
	for i, x := range goxs {
		gnoxs[i] = toExpr(x)
	}
	return
}

func toStmt(gos ast.Stmt) Stmt {
	gnos := Go2Gno(gos)
	if gnos == nil {
		return nil
	} else {
		return gnos.(Stmt)
	}
}

func toStmts(goss []ast.Stmt) (gnoss Body) {
	gnoss = make([]Stmt, len(goss))
	for i, x := range goss {
		gnoss[i] = toStmt(x)
	}
	return
}

func toBody(body *ast.BlockStmt) Body {
	if body == nil {
		return nil
	}
	return toStmts(body.List)
}

func toSimp(gos ast.Stmt) Stmt {
	gnos := Go2Gno(gos)
	if gnos == nil {
		return nil
	} else {
		return gnos.(SimpleStmt).(Stmt)
	}
}

func toDecl(god ast.Decl) Decl {
	gnod := Go2Gno(god)
	if gnod == nil {
		return nil
	} else {
		return gnod.(Decl)
	}
}

func toDecls(gd *ast.GenDecl) (ds Decls) {
	ds = make([]Decl, 0, len(gd.Specs))
	/*
		Within a parenthesized const declaration list the
		expression list may be omitted from any but the first
		ConstSpec. Such an empty list is equivalent to the textual
		substitution of the first preceding non-empty expression
		list and its type if any.
	*/
	var lastValues Exprs // (see Go iota spec above)
	var lastType Expr    // (see Go iota spec above)
	for si, s := range gd.Specs {

		switch s := s.(type) {
		case *ast.TypeSpec:
			name := toName(s.Name)
			tipe := toExpr(s.Type)
			alias := s.Assign != 0
			ds = append(ds, &TypeDecl{
				NameExpr: NameExpr{Name: name},
				Type:     tipe,
				IsAlias:  alias,
			})
		case *ast.ValueSpec:
			if gd.Tok == token.CONST {
				var values Exprs
				var tipe Expr
				if s.Values == nil {
					values = copyExprs(lastValues)
				} else {
					values = toExprs(s.Values)
					lastValues = values
				}
				if s.Type == nil {
					tipe = lastType
				} else {
					tipe = toExpr(s.Type)
					lastType = tipe
				}
				for i, id := range s.Names {
					name := toName(id)
					valu := values[i]
					cd := &ValueDecl{
						NameExpr: NameExpr{Name: name},
						Type:     tipe,
						Value:    valu,
						Const:    true,
					}
					cd.SetAttribute(ATTR_IOTA, si)
					ds = append(ds, cd)
				}
			} else {
				for i, id := range s.Names {
					name := toName(id)
					valu := Expr(nil)
					if s.Values != nil {
						valu = toExpr(s.Values[i])
					}
					tipe := toExpr(s.Type)
					ds = append(ds, &ValueDecl{
						NameExpr: NameExpr{Name: name},
						Type:     tipe,
						Value:    valu,
						Const:    false,
					})
				}
			}
		case *ast.ImportSpec:
			path, err := strconv.Unquote(s.Path.Value)
			if err != nil {
				panic("unexpected import spec path type")
			}
			ds = append(ds, &ImportDecl{
				NameExpr: NameExpr{Name: toName(s.Name)},
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

func toSimpleDecls(gd *ast.GenDecl) (sds Decls) {
	ds := toDecls(gd)
	sds = make([]Decl, len(ds))
	for i, d := range ds {
		sds[i] = d.(SimpleDecl).(Decl)
	}
	return
}

func toFieldsFromList(fl *ast.FieldList) (ftxs []FieldTypeExpr) {
	if fl == nil {
		return nil
	} else {
		ftxs = toFields(fl.List...)
		return
	}
}

func toFields(fs ...*ast.Field) (ftxs []FieldTypeExpr) {
	if len(fs) == 0 {
		return nil
	}
	ftxs = make([]FieldTypeExpr, 0, len(fs)) // may grow longer
	for _, f := range fs {
		if len(f.Names) == 0 {
			// a single unnamed field w/ type
			ftxs = append(ftxs, FieldTypeExpr{
				Name: "",
				Type: toExpr(f.Type),
				Tag:  toExpr(f.Tag),
			})
		} else {
			// one or more named fields
			for _, n := range f.Names {
				ftxs = append(ftxs, FieldTypeExpr{
					Name: toName(n),
					Type: toExpr(f.Type),
					Tag:  toExpr(f.Tag),
				})
			}
		}
	}
	return
}

func toKeyValueExprs(elts []ast.Expr) (kvxs KeyValueExprs) {
	kvxs = make([]KeyValueExpr, len(elts))
	for i, x := range elts {
		if kvx, ok := x.(*ast.KeyValueExpr); ok {
			kvxs[i] = *Go2Gno(kvx).(*KeyValueExpr)
		} else {
			kvxs[i] = KeyValueExpr{
				Key:   nil,
				Value: toExpr(x),
			}
		}
	}
	return
}

func toClauses(csz []ast.Stmt) []SwitchClauseStmt {
	res := make([]SwitchClauseStmt, len(csz))
	for i, cs := range csz {
		res[i] = toSwitchClauseStmt(cs.(*ast.CaseClause))
	}
	return res
}

func toSwitchClauseStmt(cc *ast.CaseClause) SwitchClauseStmt {
	return SwitchClauseStmt{
		Cases: toExprs(cc.List),
		Body:  toStmts(cc.Body),
	}
}
