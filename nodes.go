package gno

import (
	"fmt"
	"reflect"
	rdebug "runtime/debug"
	"strconv"
)

//----------------------------------------
// Primitives

type Word int

const (
	// Special words
	ILLEGAL Word = iota

	// Names and basic type literals
	// (these words stand for classes of literals)
	NAME   // main
	INT    // 12345
	FLOAT  // 123.45
	IMAG   // 123.45i
	CHAR   // 'a'
	STRING // "abc"

	// Operators and delimiters
	ADD // +
	SUB // -
	MUL // *
	QUO // /
	REM // %

	BAND     // &
	BOR      // |
	XOR      // ^
	SHL      // <<
	SHR      // >>
	BAND_NOT // &^

	ADD_ASSIGN      // +=
	SUB_ASSIGN      // -=
	MUL_ASSIGN      // *=
	QUO_ASSIGN      // /=
	REM_ASSIGN      // %=
	BAND_ASSIGN     // &=
	BOR_ASSIGN      // |=
	XOR_ASSIGN      // ^=
	SHL_ASSIGN      // <<=
	SHR_ASSIGN      // >>=
	BAND_NOT_ASSIGN // &^=

	LAND  // &&
	LOR   // ||
	ARROW // <-
	INC   // ++
	DEC   // --

	EQL    // ==
	LSS    // <
	GTR    // >
	ASSIGN // =
	NOT    // !

	NEQ    // !=
	LEQ    // <=
	GEQ    // >=
	DEFINE // :=

	// Keywords
	BREAK
	CASE
	CHAN
	CONST
	CONTINUE

	DEFAULT
	DEFER
	ELSE
	FALLTHROUGH
	FOR

	FUNC
	GO
	GOTO
	IF
	IMPORT

	INTERFACE
	MAP
	PACKAGE
	RANGE
	RETURN

	SELECT
	STRUCT
	SWITCH
	TYPE
	VAR
)

type Name string

//----------------------------------------
// Attributes
// All nodes have attributes for general analysis purposes.
// Exported Attribute fields like Loc and Label are persisted
// even after preprocessing.  Temporary attributes (e.g. those
// for preprocessing) are stored in .data.

type Location struct {
	PkgPath string
	File    string
	Line    int
}

type Attributes struct {
	Loc   Location
	Label Name
	data  map[interface{}]interface{} // not persisted
}

func (a *Attributes) GetLocation() Location {
	return a.Loc
}

func (a *Attributes) SetLocation(loc Location) {
	a.Loc = loc
}

func (a *Attributes) GetAttribute(key interface{}) interface{} {
	return a.data[key]
}

func (a *Attributes) SetAttribute(key interface{}, value interface{}) {
	if a.data == nil {
		a.data = make(map[interface{}]interface{})
	}
	a.data[key] = value
}

//----------------------------------------
// Node

type Node interface {
	assertNode()
	String() string
	Copy() Node
	GetLocation() Location
	SetLocation(Location)
	GetAttribute(key interface{}) interface{}
	SetAttribute(key interface{}, value interface{})
}

// non-pointer receiver to help make immutable.
func (_ *NameExpr) assertNode()          {}
func (_ *BasicLitExpr) assertNode()      {}
func (_ *BinaryExpr) assertNode()        {}
func (_ *CallExpr) assertNode()          {}
func (_ *IndexExpr) assertNode()         {}
func (_ *SelectorExpr) assertNode()      {}
func (_ *SliceExpr) assertNode()         {}
func (_ *StarExpr) assertNode()          {}
func (_ *RefExpr) assertNode()           {}
func (_ *TypeAssertExpr) assertNode()    {}
func (_ *UnaryExpr) assertNode()         {}
func (_ *CompositeLitExpr) assertNode()  {}
func (_ *KeyValueExpr) assertNode()      {}
func (_ *FuncLitExpr) assertNode()       {}
func (_ *constExpr) assertNode()         {}
func (_ *FieldTypeExpr) assertNode()     {}
func (_ *ArrayTypeExpr) assertNode()     {}
func (_ *SliceTypeExpr) assertNode()     {}
func (_ *InterfaceTypeExpr) assertNode() {}
func (_ *ChanTypeExpr) assertNode()      {}
func (_ *FuncTypeExpr) assertNode()      {}
func (_ *MapTypeExpr) assertNode()       {}
func (_ *StructTypeExpr) assertNode()    {}
func (_ *constTypeExpr) assertNode()     {}
func (_ *AssignStmt) assertNode()        {}
func (_ *BlockStmt) assertNode()         {}
func (_ *BranchStmt) assertNode()        {}
func (_ *DeclStmt) assertNode()          {}
func (_ *DeferStmt) assertNode()         {}
func (_ *ExprStmt) assertNode()          {}
func (_ *ForStmt) assertNode()           {}
func (_ *GoStmt) assertNode()            {}
func (_ *IfStmt) assertNode()            {}
func (_ *IfCaseStmt) assertNode()        {}
func (_ *IncDecStmt) assertNode()        {}
func (_ *LabeledStmt) assertNode()       {}
func (_ *RangeStmt) assertNode()         {}
func (_ *ReturnStmt) assertNode()        {}
func (_ *PanicStmt) assertNode()         {}
func (_ *SelectStmt) assertNode()        {}
func (_ *SelectCaseStmt) assertNode()    {}
func (_ *SendStmt) assertNode()          {}
func (_ *SwitchStmt) assertNode()        {}
func (_ *SwitchClauseStmt) assertNode()  {}
func (_ *EmptyStmt) assertNode()         {}
func (_ *bodyStmt) assertNode()          {}
func (_ *FuncDecl) assertNode()          {}
func (_ *ImportDecl) assertNode()        {}
func (_ *ValueDecl) assertNode()         {}
func (_ *TypeDecl) assertNode()          {}
func (_ *FileNode) assertNode()          {}
func (_ *PackageNode) assertNode()       {}

var _ Node = &NameExpr{}
var _ Node = &BasicLitExpr{}
var _ Node = &BinaryExpr{}
var _ Node = &CallExpr{}
var _ Node = &IndexExpr{}
var _ Node = &SelectorExpr{}
var _ Node = &SliceExpr{}
var _ Node = &StarExpr{}
var _ Node = &RefExpr{}
var _ Node = &TypeAssertExpr{}
var _ Node = &UnaryExpr{}
var _ Node = &CompositeLitExpr{}
var _ Node = &KeyValueExpr{}
var _ Node = &FuncLitExpr{}
var _ Node = &constExpr{}
var _ Node = &FieldTypeExpr{}
var _ Node = &ArrayTypeExpr{}
var _ Node = &SliceTypeExpr{}
var _ Node = &InterfaceTypeExpr{}
var _ Node = &ChanTypeExpr{}
var _ Node = &FuncTypeExpr{}
var _ Node = &MapTypeExpr{}
var _ Node = &StructTypeExpr{}
var _ Node = &constTypeExpr{}
var _ Node = &AssignStmt{}
var _ Node = &BlockStmt{}
var _ Node = &BranchStmt{}
var _ Node = &DeclStmt{}
var _ Node = &DeferStmt{}
var _ Node = &ExprStmt{}
var _ Node = &ForStmt{}
var _ Node = &GoStmt{}
var _ Node = &IfStmt{}
var _ Node = &IfCaseStmt{}
var _ Node = &IncDecStmt{}
var _ Node = &LabeledStmt{}
var _ Node = &RangeStmt{}
var _ Node = &ReturnStmt{}
var _ Node = &PanicStmt{}
var _ Node = &SelectStmt{}
var _ Node = &SelectCaseStmt{}
var _ Node = &SendStmt{}
var _ Node = &SwitchStmt{}
var _ Node = &SwitchClauseStmt{}
var _ Node = &EmptyStmt{}
var _ Node = &bodyStmt{}
var _ Node = &FuncDecl{}
var _ Node = &ImportDecl{}
var _ Node = &ValueDecl{}
var _ Node = &TypeDecl{}
var _ Node = &FileNode{}
var _ Node = &PackageNode{}

//----------------------------------------
// Expr
//
// expressions generally have no side effects on the caller's context,
// except for channel blocks, type assertions, and panics.

type Expr interface {
	Node
	assertExpr()
}

type Exprs []Expr

// non-pointer receiver to help make immutable.
func (*NameExpr) assertExpr()         {}
func (*BasicLitExpr) assertExpr()     {}
func (*BinaryExpr) assertExpr()       {}
func (*CallExpr) assertExpr()         {}
func (*IndexExpr) assertExpr()        {}
func (*SelectorExpr) assertExpr()     {}
func (*SliceExpr) assertExpr()        {}
func (*StarExpr) assertExpr()         {}
func (*RefExpr) assertExpr()          {}
func (*TypeAssertExpr) assertExpr()   {}
func (*UnaryExpr) assertExpr()        {}
func (*CompositeLitExpr) assertExpr() {}
func (*KeyValueExpr) assertExpr()     {}
func (*FuncLitExpr) assertExpr()      {}
func (*constExpr) assertExpr()        {}

var _ Expr = &NameExpr{}
var _ Expr = &BasicLitExpr{}
var _ Expr = &BinaryExpr{}
var _ Expr = &CallExpr{}
var _ Expr = &IndexExpr{}
var _ Expr = &SelectorExpr{}
var _ Expr = &SliceExpr{}
var _ Expr = &StarExpr{}
var _ Expr = &RefExpr{}
var _ Expr = &TypeAssertExpr{}
var _ Expr = &UnaryExpr{}
var _ Expr = &CompositeLitExpr{}
var _ Expr = &KeyValueExpr{}
var _ Expr = &FuncLitExpr{}
var _ Expr = &constExpr{}

type NameExpr struct {
	Attributes
	// TODO rename .Path's to .ValuePaths.
	Path ValuePath // set by preprocessor.
	Name
}

type NameExprs []NameExpr

type BasicLitExpr struct {
	Attributes
	// INT, FLOAT, IMAG, CHAR, or STRING
	Kind Word
	// literal string; e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f', "foo"
	// or `\m\n\o`
	Value string
}

type BinaryExpr struct { // (Left Op Right)
	Attributes
	Left  Expr // left operand
	Op    Word // operator
	Right Expr // right operand
}

type CallExpr struct { // Func(Args<Varg?...>)
	Attributes
	Func    Expr  // function expression
	Args    Exprs // function arguments, if any.
	Varg    bool  // if true, final arg is variadic.
	NumArgs int   // len(Args) or len(Args[0].Results)
}

type IndexExpr struct { // X[Index]
	Attributes
	X     Expr // expression
	Index Expr // index expression
	HasOK bool // if true, is form: `value, ok := <X>[<Key>]
}

type SelectorExpr struct { // X.Sel
	Attributes
	X    Expr      // expression
	Path ValuePath // set by preprocessor.
	Sel  Name      // field selector
}

type SliceExpr struct { // X[Low:High:Max]
	Attributes
	X    Expr // expression
	Low  Expr // begin of slice range; or nil
	High Expr // end of slice range; or nil
	Max  Expr // maximum capacity of slice; or nil; added in Go 1.2
}

// A StarExpr node represents an expression of the form
// "*" Expression.  Semantically it could be a unary "*"
// expression, or a pointer type.
type StarExpr struct { // *X
	Attributes
	X Expr // operand
}

type RefExpr struct { // &X
	Attributes
	X Expr // operand
}

type TypeAssertExpr struct { // X.(Type)
	Attributes
	X     Expr // expression.
	Type  Expr // asserted type, never nil.
	HasOK bool // if true, is form: `_, ok := <X>.(<Type>)`.
}

// A UnaryExpr node represents a unary expression. Unary
// "*" expressions (dereferencing and pointer-types) are
// represented with StarExpr nodes.  Unary & expressions
// (referencing) are represented with RefExpr nodes.
type UnaryExpr struct { // (Op X)
	Attributes
	X  Expr // operand
	Op Word // operator
}

// MyType{<key>:<value>} struct, array, slice, and map
// expressions.
type CompositeLitExpr struct {
	Attributes
	Type Expr          // literal type; or nil
	Elts KeyValueExprs // list of struct fields; if any
}

// Returns true if any elements are keyed.
// Panics if inconsistent.
func (clx *CompositeLitExpr) IsKeyed() bool {
	if len(clx.Elts) == 0 {
		return false
	} else if clx.Elts[0].Key == nil {
		for i := 1; i < len(clx.Elts); i++ {
			if clx.Elts[i].Key != nil {
				panic("mixed keyed and unkeyed elements")
			}
		}
		return false
	} else {
		for i := 1; i < len(clx.Elts); i++ {
			if clx.Elts[i].Key == nil {
				panic("mixed keyed and unkeyed elements")
			}
		}
		return true
	}
}

// A KeyValueExpr represents a single key-value pair in
// struct, array, slice, and map expressions.
type KeyValueExpr struct {
	Attributes
	Key   Expr // or nil
	Value Expr // never nil
}

type KeyValueExprs []KeyValueExpr

// A FuncLitExpr node represents a function literal.  Here one
// can reference statements from an expression, which
// completes the procedural circle.
type FuncLitExpr struct {
	Attributes
	StaticBlock
	Type FuncTypeExpr // function type
	Body              // function body
}

// The preprocessor replaces const expressions
// with *constExpr nodes.
type constExpr struct {
	Attributes
	Source Expr // (preprocessed) source of this value.
	TypedValue
}

//----------------------------------------
// Type(Expressions)
//
// In Go, Type expressions can be evaluated immediately
// without invoking the stack machine.  Exprs in type
// expressions are const (as in array len expr or map key type
// expr) or refer to an exposed symbol (with any pointer
// indirections).  this makes for more optimal performance.
//
// In Gno, type expressions are evaluated on the stack, with
// continuation opcodes, so the Gno VM could support types as
// first class objects.

type TypeExpr interface {
	Expr
	assertTypeExpr()
}

// non-pointer receiver to help make immutable.
func (_ *FieldTypeExpr) assertTypeExpr()     {}
func (_ *ArrayTypeExpr) assertTypeExpr()     {}
func (_ *SliceTypeExpr) assertTypeExpr()     {}
func (_ *InterfaceTypeExpr) assertTypeExpr() {}
func (_ *ChanTypeExpr) assertTypeExpr()      {}
func (_ *FuncTypeExpr) assertTypeExpr()      {}
func (_ *MapTypeExpr) assertTypeExpr()       {}
func (_ *StructTypeExpr) assertTypeExpr()    {}
func (_ *constTypeExpr) assertTypeExpr()     {}

func (_ *FieldTypeExpr) assertExpr()     {}
func (_ *ArrayTypeExpr) assertExpr()     {}
func (_ *SliceTypeExpr) assertExpr()     {}
func (_ *InterfaceTypeExpr) assertExpr() {}
func (_ *ChanTypeExpr) assertExpr()      {}
func (_ *FuncTypeExpr) assertExpr()      {}
func (_ *MapTypeExpr) assertExpr()       {}
func (_ *StructTypeExpr) assertExpr()    {}
func (_ *constTypeExpr) assertExpr()     {}

var _ TypeExpr = &FieldTypeExpr{}
var _ TypeExpr = &ArrayTypeExpr{}
var _ TypeExpr = &SliceTypeExpr{}
var _ TypeExpr = &InterfaceTypeExpr{}
var _ TypeExpr = &ChanTypeExpr{}
var _ TypeExpr = &FuncTypeExpr{}
var _ TypeExpr = &MapTypeExpr{}
var _ TypeExpr = &StructTypeExpr{}
var _ TypeExpr = &constTypeExpr{}

type FieldTypeExpr struct {
	Attributes
	Name
	Type Expr

	// Currently only BasicLitExpr allowed.
	// NOTE: In Go, only struct fields can have tags.
	Tag Expr
}

type FieldTypeExprs []FieldTypeExpr

func (ftxz FieldTypeExprs) IsNamed() bool {
	named := false
	for i, ftx := range ftxz {
		if i == 0 {
			named = ftx.Name != ""
		} else {
			if named && ftx.Name == "" {
				panic("[]FieldTypeExpr has inconsistent namedness (starts named)")
			} else if !named && ftx.Name != "" {
				panic("[]FieldTypeExpr has inconsistent namedness (starts unnamed)")
			}
		}
	}
	return named
}

type ArrayTypeExpr struct {
	Attributes
	Len Expr // if nil, variadic array lit
	Elt Expr // element type
}

type SliceTypeExpr struct {
	Attributes
	Elt Expr // element type
	Vrd bool // variadic arg expression
}

type InterfaceTypeExpr struct {
	Attributes
	Methods FieldTypeExprs // list of methods
	Generic Name           // for uverse generics
}

type ChanDir int

const (
	SEND ChanDir = 1 << iota
	RECV
)
const (
	BOTH = SEND | RECV
)

type ChanTypeExpr struct {
	Attributes
	Dir   ChanDir // channel direction
	Value Expr    // value type
}

type FuncTypeExpr struct {
	Attributes
	Params  FieldTypeExprs // (incoming) parameters, if any.
	Results FieldTypeExprs // (outgoing) results, if any.
}

type MapTypeExpr struct {
	Attributes
	Key   Expr // const
	Value Expr // value type
}

type StructTypeExpr struct {
	Attributes
	Fields FieldTypeExprs // list of field declarations
}

// Like constExpr but for types.
type constTypeExpr struct {
	Attributes
	Source Expr
	Type   Type
}

//----------------------------------------
// Stmt
//
// statements generally have side effects on the calling context.

type Stmt interface {
	Node
	assertStmt()
}

type Body []Stmt

func (ss Body) GetBody() Body {
	return ss
}

func (ss Body) GetLabeledStmt(label Name) (Stmt, int) {
	for i, stmt := range ss {
		if ls, ok := stmt.(*LabeledStmt); ok {
			return ls.Stmt, i
		}
	}
	return nil, -1
}

//----------------------------------------

// non-pointer receiver to help make immutable.
func (*AssignStmt) assertStmt()       {}
func (*BlockStmt) assertStmt()        {}
func (*BranchStmt) assertStmt()       {}
func (*DeclStmt) assertStmt()         {}
func (*DeferStmt) assertStmt()        {}
func (*EmptyStmt) assertStmt()        {} // useful for _ctif
func (*ExprStmt) assertStmt()         {}
func (*ForStmt) assertStmt()          {}
func (*GoStmt) assertStmt()           {}
func (*IfStmt) assertStmt()           {}
func (*IfCaseStmt) assertStmt()       {}
func (*IncDecStmt) assertStmt()       {}
func (*LabeledStmt) assertStmt()      {}
func (*RangeStmt) assertStmt()        {}
func (*ReturnStmt) assertStmt()       {}
func (*PanicStmt) assertStmt()        {}
func (*SelectStmt) assertStmt()       {}
func (*SelectCaseStmt) assertStmt()   {}
func (*SendStmt) assertStmt()         {}
func (*SwitchStmt) assertStmt()       {}
func (*SwitchClauseStmt) assertStmt() {}
func (*bodyStmt) assertStmt()         {}

var _ Stmt = &AssignStmt{}
var _ Stmt = &BlockStmt{}
var _ Stmt = &BranchStmt{}
var _ Stmt = &DeclStmt{}
var _ Stmt = &DeferStmt{}
var _ Stmt = &EmptyStmt{}
var _ Stmt = &ExprStmt{}
var _ Stmt = &ForStmt{}
var _ Stmt = &GoStmt{}
var _ Stmt = &IfStmt{}
var _ Stmt = &IfCaseStmt{}
var _ Stmt = &IncDecStmt{}
var _ Stmt = &LabeledStmt{}
var _ Stmt = &RangeStmt{}
var _ Stmt = &ReturnStmt{}
var _ Stmt = &PanicStmt{}
var _ Stmt = &SelectStmt{}
var _ Stmt = &SelectCaseStmt{}
var _ Stmt = &SendStmt{}
var _ Stmt = &SwitchStmt{}
var _ Stmt = &SwitchClauseStmt{}
var _ Stmt = &bodyStmt{}

type AssignStmt struct {
	Attributes
	Lhs Exprs
	Op  Word // assignment word (DEFINE, ASSIGN)
	Rhs Exprs
}

type BlockStmt struct {
	Attributes
	StaticBlock
	Body
}

type BranchStmt struct {
	Attributes
	Op        Word  // keyword word (BREAK, CONTINUE, GOTO, FALLTHROUGH)
	Label     Name  // label name; or empty
	Depth     uint8 // blocks to pop
	BodyIndex int   // index of statement of body
}

type DeclStmt struct {
	Attributes
	Body // (simple) ValueDecl or TypeDecl
}

type DeferStmt struct {
	Attributes
	Call CallExpr
}

// A compile artifact to use in place of nil.
// For example, _ctif() may return an empty statement.
type EmptyStmt struct {
	Attributes
}

type ExprStmt struct {
	Attributes
	X Expr
}

type ForStmt struct {
	Attributes
	StaticBlock
	Init Stmt // initialization (simple) statement; or nil
	Cond Expr // condition; or nil
	Post Stmt // post iteration (simple) statement; or nil
	Body
}

type GoStmt struct {
	Attributes
	Call CallExpr
}

// NOTE: syntactically, code may choose to chain if-else statements
// with `} else if ... {` constructions, but this is not represented
// in the logical AST.
type IfStmt struct {
	Attributes
	StaticBlock
	Init Stmt       // initialization (simple) statement; or nil
	Cond Expr       // condition; or nil
	Then IfCaseStmt // body statements
	Else IfCaseStmt // else statements
}

type IfCaseStmt struct {
	Attributes
	StaticBlock
	Body
}

type IncDecStmt struct {
	Attributes
	X  Expr
	Op Word // INC or DEC
}

type LabeledStmt struct {
	Attributes
	Label Name
	Stmt  Stmt
}

type RangeStmt struct {
	Attributes
	StaticBlock
	X          Expr // value to range over
	Key, Value Expr // Key, Value may be nil
	Op         Word // ASSIGN or DEFINE
	Body
	IsMap      bool // if X is map type
	IsString   bool // if X is string type
	IsArrayPtr bool // if X is array-pointer type
}

type ReturnStmt struct {
	Attributes
	Results Exprs // result expressions; or nil
}

type PanicStmt struct {
	Attributes
	Exception Expr // panic expression; not nil
}

type SelectStmt struct {
	Attributes
	Cases []SelectCaseStmt
}

type SelectCaseStmt struct {
	Attributes
	StaticBlock
	Comm Stmt // send or receive statement; nil means default case
	Body
}

type SendStmt struct {
	Attributes
	Chan  Expr
	Value Expr
}

// type ReceiveStmt
// is just AssignStmt with a Receive unary expression.

type SwitchStmt struct {
	Attributes
	StaticBlock
	Init         Stmt               // init (simple) stmt; or nil
	X            Expr               // tag or _.(type) expr; or nil
	IsTypeSwitch bool               // true iff X is .(type) expr
	Clauses      []SwitchClauseStmt // case clauses
	VarName      Name               // type-switched value; or ""
}

type SwitchClauseStmt struct {
	Attributes
	StaticBlock
	Cases Exprs // list of expressions or types; nil means default case
	Body
}

//----------------------------------------
// bodyStmt (persistent)

// NOTE: embedded in Block.
type bodyStmt struct {
	Attributes
	Body                       // for non-loop stmts
	BodyLen       int          // for for-continue
	NextBodyIndex int          // init:-2, cond/elem:-1, body:0..., post:n
	NumOps        int          // number of Ops, for goto
	NumStmts      int          // number of Stmts, for goto
	Cond          Expr         // for ForStmt
	Post          Stmt         // for ForStmt
	Active        Stmt         // for PopStmt()
	Key           Expr         // for RangeStmt
	Value         Expr         // for RangeStmt
	Op            Word         // for RangeStmt
	ListLen       int          // for RangeStmt only
	ListIndex     int          // for RangeStmt only
	NextItem      *MapListItem // fpr RangeStmt w/ maps only
	StrLen        int          // for RangeStmt w/ strings only
	StrIndex      int          // for RangeStmt w/ strings only
	NextRune      rune         // for RangeStmt w/ strings only
}

func (s *bodyStmt) PopActiveStmt() (as Stmt) {
	as = s.Active
	s.Active = nil
	return
}

func (s *bodyStmt) String() string {
	next := ""
	if s.NextBodyIndex < 0 {
		next = "(init)"
	} else if s.NextBodyIndex == len(s.Body) {
		next = "(end)"
	} else {
		next = s.Body[s.NextBodyIndex].String()
	}
	active := ""
	if s.Active != nil {
		if s.NextBodyIndex < 0 || s.NextBodyIndex == len(s.Body) {
			// none
		} else if s.Body[s.NextBodyIndex-1] == s.Active {
			active = "*"
		} else {
			active = fmt.Sprintf(" unexpected active: %v", s.Active)
		}
	}
	return fmt.Sprintf("bodyStmt[%d/%d/%d]=%s%s",
		s.ListLen,
		s.ListIndex,
		s.NextBodyIndex,
		next,
		active)
}

//----------------------------------------
// Simple Statement
// NOTE: SimpleStmt is not used in nodes due to itable conversion costs.
//
// These are used in if, switch, and for statements for simple
// initialization.  The only allowed types are EmptyStmt, ExprStmt,
// SendStmt, IncDecStmt, and AssignStmt.

type SimpleStmt interface {
	Stmt
	assertSimpleStmt()
}

// non-pointer receiver to help make immutable.
func (*EmptyStmt) assertSimpleStmt()  {}
func (*ExprStmt) assertSimpleStmt()   {}
func (*SendStmt) assertSimpleStmt()   {}
func (*IncDecStmt) assertSimpleStmt() {}
func (*AssignStmt) assertSimpleStmt() {}

//----------------------------------------
// Decl

type Decl interface {
	Node
	GetDeclNames() []Name
	assertDecl()
}

type Decls []Decl

// non-pointer receiver to help make immutable.
func (_ *FuncDecl) assertDecl()   {}
func (_ *ImportDecl) assertDecl() {}
func (_ *ValueDecl) assertDecl()  {}
func (_ *TypeDecl) assertDecl()   {}

var _ Decl = &FuncDecl{}
var _ Decl = &ImportDecl{}
var _ Decl = &ValueDecl{}
var _ Decl = &TypeDecl{}

type FuncDecl struct {
	Attributes
	StaticBlock
	NameExpr
	IsMethod bool
	Recv     FieldTypeExpr // receiver (if method); or empty (if function)
	Type     FuncTypeExpr  // function signature: parameters and results
	Body                   // function body; or empty for external (non-Go) function
}

func (fd *FuncDecl) GetDeclNames() []Name {
	return []Name{fd.NameExpr.Name}
}

type ImportDecl struct {
	Attributes
	NameExpr // local package name, or ".". required.
	PkgPath  string
}

func (id *ImportDecl) GetDeclNames() []Name {
	if id.NameExpr.Name == "." {
		return nil // ignore
	} else {
		return []Name{id.NameExpr.Name}
	}
}

type ValueDecl struct {
	Attributes
	NameExprs
	Type   Expr  // value type; or nil
	Values Exprs // initial value; or nil (unless const).
	Const  bool
}

func (vd *ValueDecl) GetDeclNames() []Name {
	ns := make([]Name, 0, len(vd.NameExprs))
	for _, nx := range vd.NameExprs {
		if nx.Name == "_" {
			// ignore
		} else {
			ns = append(ns, nx.Name)
		}
	}
	return ns
}

type TypeDecl struct {
	Attributes
	NameExpr
	Type    Expr // Name, SelectorExpr, StarExpr, or XxxTypes
	IsAlias bool // type alias since Go 1.9
}

func (td *TypeDecl) GetDeclNames() []Name {
	if td.NameExpr.Name == "_" {
		return nil // ignore
	} else {
		return []Name{td.NameExpr.Name}
	}
}

func HasDeclName(d Decl, n2 Name) bool {
	ns := d.GetDeclNames()
	for _, n := range ns {
		if n == n2 {
			return true
		}
	}
	return false
}

//----------------------------------------
// SimpleDeclStmt
//
// These are elements of DeclStmt, and get pushed to m.Stmts.

type SimpleDeclStmt interface {
	Decl
	Stmt
	assertSimpleDeclStmt()
}

// not used to avoid itable costs.
// type SimpleDeclStmts []SimpleDeclStmt

func (_ *ValueDecl) assertSimpleDeclStmt() {}
func (_ *TypeDecl) assertSimpleDeclStmt()  {}

func (_ *ValueDecl) assertStmt() {}
func (_ *TypeDecl) assertStmt()  {}

var _ SimpleDeclStmt = &ValueDecl{}
var _ SimpleDeclStmt = &TypeDecl{}

//----------------------------------------
// *FileSet

type FileSet struct {
	Files []*FileNode
}

func (fs *FileSet) AddFiles(fns ...*FileNode) {
	fs.Files = append(fs.Files, fns...)
}

func (fs *FileSet) GetFileByName(n Name) *FileNode {
	for _, fn := range fs.Files {
		if fn.Name == n {
			return fn
		}
	}
	return nil
}

// Returns a pointer to the file body decl (as well as
// the *FileNode which contains it) that declares n
// for the associated package with *FileSet.  Does not
// work for import decls which are for the file level.
// The file body decl can be replaced by reference
// assignment.
// TODO move to package?
func (fs *FileSet) GetDeclFor(n Name) (*FileNode, *Decl) {
	// XXX index to bound to linear time.
	for _, fn := range fs.Files {
		for i, dn := range fn.Decls {
			if _, isImport := dn.(*ImportDecl); isImport {
				continue
			}
			if HasDeclName(dn, n) {
				// found the decl that declares n.
				return fn, &fn.Decls[i]
			}
		}
	}
	panic(fmt.Sprintf(
		"name %s not defined in fileset %v",
		n, fs))
}

//----------------------------------------
// FileNode, & PackageNode

type FileNode struct {
	Attributes
	StaticBlock
	Name
	PkgName Name
	Decls
}

type PackageNode struct {
	Attributes
	StaticBlock
	PkgPath string
	PkgName Name
	*FileSet
}

func NewPackageNode(name Name, path string, fset *FileSet) *PackageNode {
	pn := &PackageNode{
		PkgPath: path,
		PkgName: Name(name),
		FileSet: fset,
	}
	pn.InitStaticBlock(pn, nil)
	return pn
}

func (pn *PackageNode) NewPackage(rlm *Realm) *PackageValue {
	if IsRealmPath(pn.PkgPath) {
		if rlm == nil {
			panic("should not happen")
		}
	} else {
		if rlm != nil {
			panic("should not happen")
		}
	}
	pv := &PackageValue{
		Block: Block{
			Source: pn,
		},
		PkgName:    pn.PkgName,
		PkgPath:    pn.PkgPath,
		FNames:     nil,
		FBlocks:    nil,
		fBlocksMap: make(map[Name]*Block),
	}
	if rlm != nil {
		pv.SetRealm(rlm)
	}
	pn.PrepareNewValues(pv)
	return pv
}

// Prepares new func and method values by attaching the proper file block.
// Returns a slice of new PackageValue.Values.
// After return, *PackageNode.Values and *PackageValue.Values have the same
// length.
// TODO split logic and/or name resulting function(s) better. PrepareNewValues?
func (pn *PackageNode) PrepareNewValues(pv *PackageValue) []TypedValue {
	if pv.Source != pn {
		panic("PackageNode.PrepareNewValues() package mismatch")
	}
	pvl := len(pv.Values)
	pnl := len(pn.Values)
	if pvl < pnl {
		// XXX: deep copy heap values
		nvs := make([]TypedValue, pnl-pvl)
		copy(nvs, pn.Values[pvl:pnl])
		for _, nv := range nvs {
			if nv.IsUndefined() {
				continue
			}
			switch nv.T.Kind() {
			case FuncKind:
				// If package-level FuncLit function, value is nil,
				// and the closure will be set at run-time.
				if nv.V == nil {
					// nothing to do
				} else {
					// Set function closure for function declarations.
					switch fv := nv.V.(type) {
					case *FuncValue:
						fv.PkgPath = pv.PkgPath
						fv.pkg = pv
						if fv.Closure != nil {
							panic("expected nil closure for static func")
						}
						if fv.FileName == "" {
							// Allow
							// m.RunDeclaration(FuncD(...))
							// without any file nodes, as long
							// as it uses no imports.
							continue
						}
						fb := pv.fBlocksMap[fv.FileName]
						if fb == nil {
							panic("should not happen")
						}
						fv.Closure = fb
					case *nativeValue:
						// do nothing for go native functions.
					default:
						panic("should not happen")
					}
				}
			case TypeKind:
				nt := nv.GetType()
				if dt, ok := nt.(*DeclaredType); ok {
					for i := 0; i < len(dt.Methods); i++ {
						mv := dt.Methods[i].V.(*FuncValue)
						if mv.Closure != nil {
							// This happens with alias declarations.
						}
						// set mv.pkg.
						mv.PkgPath = pv.PkgPath
						mv.pkg = pv
						// set mv.Closure.
						fn, _ := pn.GetDeclFor(dt.Name)
						fb := pv.fBlocksMap[fn.Name]
						if fb == nil {
							panic("should not happen")
						}
						mv.Closure = fb
					}
				}
			default:
				// already shallowed copied.
			}
		}
		pv.Values = append(pv.Values, nvs...)
		return pv.Values[pvl:]
	} else if pvl > pnl {
		panic("package size error")
	} else {
		// nothing to do
		return nil
	}
}

//----------------------------------------
// BlockNode

// Nodes that create their own scope satisfy this interface.
type BlockNode interface {
	Node
	InitStaticBlock(BlockNode, BlockNode)
	GetStaticBlock() *StaticBlock

	// StaticBlock promoted methods
	GetBlockNames() []Name
	GetExternNames() []Name
	GetNumNames() uint16
	GetParentNode(Store) BlockNode
	GetPathForName(Store, Name) ValuePath
	GetLocalIndex(Name) (uint16, bool)
	GetValueRef(Store, Name) *TypedValue
	GetStaticTypeOf(Store, Name) Type
	GetStaticTypeOfAt(Store, ValuePath) Type
	Define(Name, TypedValue)
	Define2(Name, Type, TypedValue)
	GetBody() Body
}

//----------------------------------------
// StaticBlock

// Embed in node to make it a BlockNode.
type StaticBlock struct {
	Block
	Types    []Type
	NumNames uint16
	Names    []Name
	Externs  []Name
}

// Implements BlockNode
func (sb *StaticBlock) InitStaticBlock(source BlockNode, parent BlockNode) {
	if sb.Names != nil {
		panic("StaticBlock.Names already initalized")
	}
	if parent == nil {
		sb.Block = Block{
			Source: source,
			Values: nil,
			Parent: nil,
		}
	} else {
		sb.Block = Block{
			Source: source,
			Values: nil,
			Parent: parent.GetStaticBlock().GetBlock(),
		}
	}
	sb.NumNames = 0
	sb.Names = make([]Name, 0, 16)
	sb.Externs = make([]Name, 0, 16)
	return
}

// Implements BlockNode.
func (sb *StaticBlock) GetStaticBlock() *StaticBlock {
	return sb
}

// Does not implement BlockNode to prevent confusion.
// To get the static *Block, call Blocknode.GetStaticBlock().GetBlock().
func (sb *StaticBlock) GetBlock() *Block {
	return &sb.Block
}

// Implements BlockNode.
func (sb *StaticBlock) GetBlockNames() (ns []Name) {
	return sb.Names // copy?
}

// Implements BlockNode.
func (sb *StaticBlock) GetExternNames() (ns []Name) {
	return sb.Externs // copy?
}

func (sb *StaticBlock) addExternName(n Name) {
	for _, extern := range sb.Externs {
		if extern == n {
			return
		}
	}
	sb.Externs = append(sb.Externs, n)
}

// Implements BlockNode.
func (sb *StaticBlock) GetNumNames() (nn uint16) {
	return sb.NumNames
}

// Implements BlockNode.
func (sb *StaticBlock) GetParentNode(store Store) BlockNode {
	pblock := sb.Block.GetParent(store)
	if pblock == nil {
		return nil
	} else {
		return pblock.Source
	}
}

// Implements BlockNode.
// As a side effect, notes externally defined names.
func (sb *StaticBlock) GetPathForName(store Store, n Name) ValuePath {
	if debug {
		if n == "_" {
			panic("should not happen")
		}
	}
	gen := 1
	if idx, ok := sb.GetLocalIndex(n); ok {
		return NewValuePathBlock(uint8(gen), idx, n)
	} else {
		if !isFile(sb.Source) {
			sb.GetStaticBlock().addExternName(n)
		}
		gen++
		bp := sb.GetParentNode(store)
		for bp != nil {
			if idx, ok = bp.GetLocalIndex(n); ok {
				return NewValuePathBlock(uint8(gen), idx, n)
			} else {
				if !isFile(bp) {
					sb.GetStaticBlock().addExternName(n)
				}
				bp = bp.GetParentNode(store)
				gen++
				if 0xff < gen {
					panic("value path depth overflow")
				}
			}
		}
		panic(fmt.Sprintf("name %s not declared", n))
	}
}

// Implements BlockNode.
func (sb *StaticBlock) GetStaticTypeOf(store Store, n Name) Type {
	idx, ok := sb.GetLocalIndex(n)
	ts := sb.Types
	bp := sb.GetParentNode(store)
	for {
		if ok {
			return ts[idx]
		} else if bp != nil {
			idx, ok = bp.GetLocalIndex(n)
			ts = bp.GetStaticBlock().Types
			bp = bp.GetParentNode(store)
		} else {
			panic(fmt.Sprintf("name %s not declared", n))
		}
	}
}

// Implements BlockNode.
func (sb *StaticBlock) GetStaticTypeOfAt(store Store, path ValuePath) Type {
	if debug {
		if path.Type != VPBlock {
			panic("should not happen")
		}
		if path.Depth == 0 {
			panic("should not happen")
		}
	}
	for {
		if path.Depth == 1 {
			return sb.Types[path.Index]
		} else {
			sb = sb.GetParentNode(store).GetStaticBlock()
			path.Depth -= 1
		}
	}
	panic("should not happen")
}

// Implements BlockNode.
func (sb *StaticBlock) GetLocalIndex(n Name) (uint16, bool) {
	for i, name := range sb.Names {
		if name == n {
			if debug {
				nt := reflect.TypeOf(sb.Source).String()
				debug.Printf("StaticBlock(%p %v).GetLocalIndex(%s) = %v, %v\n",
					sb, nt, n, i, name)
			}
			return uint16(i), true
		}
	}
	if debug {
		nt := reflect.TypeOf(sb.Source).String()
		debug.Printf("StaticBlock(%p %v).GetLocalIndex(%s) = undefined\n",
			sb, nt, n)
	}
	return 0, false
}

// Implemented BlockNode.
// This method is too slow for runtime, but it is used
// during preprocessing to compute types.
// Returns nil if not defined.
func (sb *StaticBlock) GetValueRef(store Store, n Name) *TypedValue {
	idx, ok := sb.GetLocalIndex(n)
	bb := &sb.Block
	bp := sb.GetParentNode(store)
	for {
		if ok {
			return bb.GetPointerToInt(store, int(idx)).TV
		} else if bp != nil {
			idx, ok = bp.GetLocalIndex(n)
			bb = bp.GetStaticBlock().GetBlock()
			bp = bp.GetParentNode(store)
		} else {
			return nil
		}
	}
}

// Implements BlockNode
// Statically declares a name definition.
// At runtime, use *Block.GetValueRef() etc which take path
// values, which are pre-computeed in the preprocessor.
// Once a typed value is defined, it cannot be changed.
//
// NOTE: Currently tv.V is only set when the value
// represents a Type(Value). The purpose of tv is to describe
// the invariant of a named value, at the minimum its type,
// but also sometimes the typeval value; but we could go
// further and store preprocessed constant results here
// too.  See "anyValue()" and "asValue()" for usage.
func (sb *StaticBlock) Define(n Name, tv TypedValue) {
	sb.Define2(n, tv.T, tv)
}

func (sb *StaticBlock) Define2(n Name, st Type, tv TypedValue) {
	if debug {
		debug.Printf(
			"StaticBlock.Define(%s, %v)\n",
			n, tv)
	}
	// TODO check that tv.T implements t.
	if len(n) == 0 {
		panic("name cannot be zero")
	}
	if int(sb.NumNames) != len(sb.Names) {
		panic("StaticBlock.NumNames and len(.Names) mismatch")
	}
	if (1<<16 - 1) < sb.NumNames {
		panic("too many variables in block")
	}
	if tv.T == nil && tv.V != nil {
		panic("StaticBlock.Define() requires .T if .V is set")
	}
	idx, exists := sb.GetLocalIndex(n)
	if exists {
		// Is re-defining.
		old := sb.Block.Values[idx]
		if !old.IsUndefined() {
			if tv.T != old.T {
				panic(fmt.Sprintf(
					"StaticBlock.Define() cannot change .T; was %v, new %v",
					old.T, tv.T))
			}
			if tv.V != old.V {
				panic("StaticBlock.Define() cannot change .V")
			}
		}
		sb.Block.Values[idx] = tv
		sb.Types[idx] = st
	} else {
		// The general case without re-definition.
		sb.Names = append(sb.Names, n)
		sb.NumNames++
		sb.Block.Values = append(sb.Block.Values, tv)
		sb.Types = append(sb.Types, st)
	}
}

// Implements BlockNode
func (sb *StaticBlock) SetStaticBlock(osb StaticBlock) {
	*sb = osb
}

var _ BlockNode = &FuncLitExpr{}
var _ BlockNode = &BlockStmt{}
var _ BlockNode = &ForStmt{}
var _ BlockNode = &IfStmt{} // faux block node
var _ BlockNode = &IfCaseStmt{}
var _ BlockNode = &RangeStmt{}
var _ BlockNode = &SelectCaseStmt{}
var _ BlockNode = &SwitchStmt{} // faux block node
var _ BlockNode = &SwitchClauseStmt{}
var _ BlockNode = &FuncDecl{}
var _ BlockNode = &FileNode{}
var _ BlockNode = &PackageNode{}

func (ifs *IfStmt) GetBody() Body {
	panic("IfStmt has no body (but .Then and .Else do)")
}

func (ifs *SwitchStmt) GetBody() Body {
	panic("SwitchStmt has no body (but its cases do)")
}

func (fn *FileNode) GetBody() Body {
	panic("FileNode has no body (but it does have .Decls)")
}

func (pn *PackageNode) GetBody() Body {
	panic("PackageNode has no body")
}

//----------------------------------------
// Value Path

// A relative pointer to a TypedValue value
//  (a) a Block scope var or const
//  (b) a StructValue field
//  (c) a DeclaredType method
//  (d) a PackageNode declaration
//
// Depth tells how many layers of access should be unvealed before
// arriving at the ultimate handler type.  In the case of Blocks,
// the depth tells how many layers of ancestry to ascend before
// arriving at the target block.  For other selector expr paths
// such as those for *DeclaredType methods or *StructType fields,
// see tests/selector_test.go.
type ValuePath struct {
	Type  VPType // see VPType* consts.
	Depth uint8  // see doc for ValuePath.
	Index uint16 // index of value, field, or method.
	Name  Name   // name of value, field, or method.
}

type VPType uint8

const (
	VPUverse         VPType = 0x00
	VPBlock          VPType = 0x01 // blocks and packages
	VPField          VPType = 0x02
	VPValMethod      VPType = 0x03
	VPPtrMethod      VPType = 0x04
	VPInterface      VPType = 0x05
	VPSubrefField    VPType = 0x06 // not deref type
	VPDerefField     VPType = 0x12 // 0x10 + VPField
	VPDerefValMethod VPType = 0x13 // 0x10 + VPValMethod
	VPDerefPtrMethod VPType = 0x14 // 0x10 + VPPtrMethod
	VPDerefInterface VPType = 0x15 // 0x10 + VPInterface
	VPNative         VPType = 0x20
	// 0x3X, 0x5X, 0x7X, 0x9X, 0xAX, 0xCX, 0xEX reserved.
)

func NewValuePath(t VPType, depth uint8, index uint16, n Name) ValuePath {
	if t == VPField && n == "Root" {
		rdebug.PrintStack()
	}
	vp := ValuePath{
		Type:  t,
		Depth: depth,
		Index: index,
		Name:  n,
	}
	vp.Validate()
	return vp
}

func NewValuePathUverse(index uint16, n Name) ValuePath {
	return NewValuePath(VPUverse, 0, uint16(index), n)
}

func NewValuePathBlock(depth uint8, index uint16, n Name) ValuePath {
	return NewValuePath(VPBlock, depth, index, n)
}

func NewValuePathField(depth uint8, index uint16, n Name) ValuePath {
	return NewValuePath(VPField, depth, index, n)
}

func NewValuePathValMethod(index uint16, n Name) ValuePath {
	return NewValuePath(VPValMethod, 0, index, n)
}

func NewValuePathPtrMethod(index uint16, n Name) ValuePath {
	return NewValuePath(VPPtrMethod, 0, index, n)
}

func NewValuePathInterface(n Name) ValuePath {
	return NewValuePath(VPInterface, 0, 0, n)
}

func NewValuePathSubrefField(depth uint8, index uint16, n Name) ValuePath {
	return NewValuePath(VPSubrefField, depth, index, n)
}

func NewValuePathDerefField(depth uint8, index uint16, n Name) ValuePath {
	return NewValuePath(VPDerefField, depth, index, n)
}

func NewValuePathDerefValMethod(index uint16, n Name) ValuePath {
	return NewValuePath(VPDerefValMethod, 0, index, n)
}

func NewValuePathDerefPtrMethod(index uint16, n Name) ValuePath {
	return NewValuePath(VPDerefPtrMethod, 0, index, n)
}

func NewValuePathDerefInterface(n Name) ValuePath {
	return NewValuePath(VPDerefInterface, 0, 0, n)
}

func NewValuePathNative(n Name) ValuePath {
	return NewValuePath(VPNative, 0, 0, n)
}

func (vp ValuePath) Validate() {
	switch vp.Type {
	case VPUverse:
		if vp.Depth != 0 {
			panic("uverse value path must have depth 0")
		}
	case VPBlock:
		// 0 ok ("_" blank)
	case VPField:
		if vp.Depth > 1 {
			panic("field value path must have depth 0 or 1")
		}
	case VPValMethod:
		if vp.Depth != 0 {
			panic("method value path must have depth 0")
		}
	case VPPtrMethod:
		if vp.Depth != 0 {
			panic("ptr receiver method value path must have depth 0")
		}
	case VPInterface:
		if vp.Depth != 0 {
			panic("interface method value path must have depth 0")
		}
		if vp.Name == "" {
			panic("interface value path must have name")
		}
	case VPSubrefField:
		if vp.Depth > 3 {
			panic("subref field value path must have depth 0, 1, 2, or 3")
		}
	case VPDerefField:
		if vp.Depth > 3 {
			panic("deref field value path must have depth 0, 1, 2, or 3")
		}
	case VPDerefValMethod:
		if vp.Depth != 0 {
			panic("(deref) method value path must have depth 0")
		}
	case VPDerefPtrMethod:
		if vp.Depth != 0 {
			panic("(deref) ptr receiver method value path must have depth 0")
		}
	case VPDerefInterface:
		if vp.Depth != 0 {
			panic("(deref) interface method value path must have depth 0")
		}
		if vp.Name == "" {
			panic("(deref) interface value path must have name")
		}
	case VPNative:
		if vp.Depth != 0 {
			panic("native value path must have depth 0")
		}
		if vp.Name == "" {
			panic("native value path must have name")
		}
	default:
		panic(fmt.Sprintf(
			"unexpected value path type %X",
			vp.Type))
	}
}

func (vp ValuePath) IsBlockBlankPath() bool {
	return vp.Type == VPBlock && vp.Depth == 0 && vp.Index == 0
}

func (vp ValuePath) IsDerefType() bool {
	return vp.Type&0x10 > 0
}

type ValuePather interface {
	GetPathForName(Name) ValuePath
}

//----------------------------------------
// Utility

func (blx *BasicLitExpr) GetString() string {
	str, err := strconv.Unquote(blx.Value)
	if err != nil {
		panic("error in parsing string literal: " + err.Error())
	}
	return str
}

func (blx *BasicLitExpr) GetInt() int {
	i, err := strconv.Atoi(blx.Value)
	if err != nil {
		panic(err)
	}
	return i
}

type GnoAttribute string

const (
	ATTR_PREPROCESSED GnoAttribute = "ATTR_PREPROCESSED"
	ATTR_PREDEFINED   GnoAttribute = "ATTR_PREDEFINED"
	ATTR_TYPE_VALUE   GnoAttribute = "ATTR_TYPE_VALUE"
	ATTR_TYPEOF_VALUE GnoAttribute = "ATTR_TYPEOF_VALUE"
	ATTR_LABEL        GnoAttribute = "ATTR_LABEL"
	ATTR_IOTA         GnoAttribute = "ATTR_IOTA"
)
