package gnolang

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// ----------------------------------------
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

type Names []Name

func (ns Names) Join(j string) string {
	ss := make([]string, 0, len(ns))
	for _, n := range ns {
		ss = append(ss, string(n))
	}
	return strings.Join(ss, j)
}

// ----------------------------------------
// Attributes
// All nodes have attributes for general analysis purposes.
// Exported Attribute fields like Loc and Label are persisted
// even after preprocessing.  Temporary attributes (e.g. those
// for preprocessing) are stored in .data.

type GnoAttribute string

// XXX once everything is done, convert to a uint64 bitflag.
const (
	ATTR_PREPROCESSED          GnoAttribute = "ATTR_PREPROCESSED"
	ATTR_PREPROCESS_SKIPPED    GnoAttribute = "ATTR_PREPROCESS_SKIPPED"
	ATTR_PREPROCESS_INCOMPLETE GnoAttribute = "ATTR_PREPROCESS_INCOMPLETE"
	ATTR_PREDEFINED            GnoAttribute = "ATTR_PREDEFINED"
	ATTR_TYPE_VALUE            GnoAttribute = "ATTR_TYPE_VALUE"
	ATTR_TYPEOF_VALUE          GnoAttribute = "ATTR_TYPEOF_VALUE"
	ATTR_IOTA                  GnoAttribute = "ATTR_IOTA"
	ATTR_HEAP_DEFINES          GnoAttribute = "ATTR_HEAP_DEFINES" // []Name heap items.
	ATTR_HEAP_USES             GnoAttribute = "ATTR_HEAP_USES"    // []Name heap items used.
	ATTR_SHIFT_RHS             GnoAttribute = "ATTR_SHIFT_RHS"
	ATTR_LAST_BLOCK_STMT       GnoAttribute = "ATTR_LAST_BLOCK_STMT"
	ATTR_PACKAGE_REF           GnoAttribute = "ATTR_PACKAGE_REF"
	ATTR_PACKAGE_DECL          GnoAttribute = "ATTR_PACKAGE_DECL"
	ATTR_PACKAGE_PATH          GnoAttribute = "ATTR_PACKAGE_PATH" // if name expr refers to package.
	ATTR_FIX_FROM              GnoAttribute = "ATTR_FIX_FROM"     // gno fix this version.
	ATTR_REDEFINE_NAME         GnoAttribute = "ATTR_REDEFINE_NAME"
)

// Embedded in each Node.
type Attributes struct {
	Span  // Node.Line is the start.
	Label Name
	data  map[GnoAttribute]any // not persisted
}

func (attr *Attributes) GetLabel() Name {
	return attr.Label
}

func (attr *Attributes) SetLabel(label Name) {
	attr.Label = label
}

func (attr *Attributes) HasAttribute(key GnoAttribute) bool {
	_, ok := attr.data[key]
	return ok
}

// GnoAttribute must not be user provided / arbitrary,
// otherwise will create potential exploits.
func (attr *Attributes) GetAttribute(key GnoAttribute) any {
	return attr.data[key]
}

func (attr *Attributes) SetAttribute(key GnoAttribute, value any) {
	if attr.data == nil {
		attr.data = make(map[GnoAttribute]any)
	}
	attr.data[key] = value
}

func (attr *Attributes) DelAttribute(key GnoAttribute) {
	if debug && attr.data == nil {
		panic("should not happen, attribute is expected to be non-empty.")
	}
	delete(attr.data, key)
}

func (attr *Attributes) GetAttributeKeys() []GnoAttribute {
	res := make([]GnoAttribute, 0, len(attr.data))
	for key := range attr.data {
		res = append(res, key)
	}
	return res
}

func (attr *Attributes) String() string {
	panic("should not use") // node should override Pos/Span/Location methods.
}

func (attr *Attributes) IsZero() bool {
	panic("should not use") // node should override Pos/Span/Location methods.
}

// ----------------------------------------
// Node

type Node interface {
	assertNode()
	String() string
	Copy() Node
	GetPos() Pos
	GetLine() int
	GetColumn() int
	GetSpan() Span
	SetSpan(Span) // once.
	GetLabel() Name
	SetLabel(Name)
	HasAttribute(key GnoAttribute) bool
	GetAttribute(key GnoAttribute) any
	SetAttribute(key GnoAttribute, value any)
	DelAttribute(key GnoAttribute)
}

// non-pointer receiver to help make immutable.
func (*NameExpr) assertNode()          {}
func (*BasicLitExpr) assertNode()      {}
func (*BinaryExpr) assertNode()        {}
func (*CallExpr) assertNode()          {}
func (*IndexExpr) assertNode()         {}
func (*SelectorExpr) assertNode()      {}
func (*SliceExpr) assertNode()         {}
func (*StarExpr) assertNode()          {}
func (*RefExpr) assertNode()           {}
func (*TypeAssertExpr) assertNode()    {}
func (*UnaryExpr) assertNode()         {}
func (*CompositeLitExpr) assertNode()  {}
func (*KeyValueExpr) assertNode()      {}
func (*FuncLitExpr) assertNode()       {}
func (*ConstExpr) assertNode()         {}
func (*FieldTypeExpr) assertNode()     {}
func (*ArrayTypeExpr) assertNode()     {}
func (*SliceTypeExpr) assertNode()     {}
func (*InterfaceTypeExpr) assertNode() {}
func (*ChanTypeExpr) assertNode()      {}
func (*FuncTypeExpr) assertNode()      {}
func (*MapTypeExpr) assertNode()       {}
func (*StructTypeExpr) assertNode()    {}
func (*constTypeExpr) assertNode()     {}
func (*AssignStmt) assertNode()        {}
func (*BlockStmt) assertNode()         {}
func (*BranchStmt) assertNode()        {}
func (*DeclStmt) assertNode()          {}
func (*DeferStmt) assertNode()         {}
func (*ExprStmt) assertNode()          {}
func (*ForStmt) assertNode()           {}
func (*GoStmt) assertNode()            {}
func (*IfStmt) assertNode()            {}
func (*IfCaseStmt) assertNode()        {}
func (*IncDecStmt) assertNode()        {}
func (*RangeStmt) assertNode()         {}
func (*ReturnStmt) assertNode()        {}
func (*SelectStmt) assertNode()        {}
func (*SelectCaseStmt) assertNode()    {}
func (*SendStmt) assertNode()          {}
func (*SwitchStmt) assertNode()        {}
func (*SwitchClauseStmt) assertNode()  {}
func (*EmptyStmt) assertNode()         {}
func (*bodyStmt) assertNode()          {}
func (*FuncDecl) assertNode()          {}
func (*ImportDecl) assertNode()        {}
func (*ValueDecl) assertNode()         {}
func (*TypeDecl) assertNode()          {}
func (*FileNode) assertNode()          {}
func (*PackageNode) assertNode()       {}

var (
	_ Node = &NameExpr{}
	_ Node = &BasicLitExpr{}
	_ Node = &BinaryExpr{}
	_ Node = &CallExpr{}
	_ Node = &IndexExpr{}
	_ Node = &SelectorExpr{}
	_ Node = &SliceExpr{}
	_ Node = &StarExpr{}
	_ Node = &RefExpr{}
	_ Node = &TypeAssertExpr{}
	_ Node = &UnaryExpr{}
	_ Node = &CompositeLitExpr{}
	_ Node = &KeyValueExpr{}
	_ Node = &FuncLitExpr{}
	_ Node = &ConstExpr{}
	_ Node = &FieldTypeExpr{}
	_ Node = &ArrayTypeExpr{}
	_ Node = &SliceTypeExpr{}
	_ Node = &InterfaceTypeExpr{}
	_ Node = &ChanTypeExpr{}
	_ Node = &FuncTypeExpr{}
	_ Node = &MapTypeExpr{}
	_ Node = &StructTypeExpr{}
	_ Node = &constTypeExpr{}
	_ Node = &AssignStmt{}
	_ Node = &BlockStmt{}
	_ Node = &BranchStmt{}
	_ Node = &DeclStmt{}
	_ Node = &DeferStmt{}
	_ Node = &ExprStmt{}
	_ Node = &ForStmt{}
	_ Node = &GoStmt{}
	_ Node = &IfStmt{}
	_ Node = &IfCaseStmt{}
	_ Node = &IncDecStmt{}
	_ Node = &RangeStmt{}
	_ Node = &ReturnStmt{}
	_ Node = &SelectStmt{}
	_ Node = &SelectCaseStmt{}
	_ Node = &SendStmt{}
	_ Node = &SwitchStmt{}
	_ Node = &SwitchClauseStmt{}
	_ Node = &EmptyStmt{}
	_ Node = &bodyStmt{}
	_ Node = &FuncDecl{}
	_ Node = &ImportDecl{}
	_ Node = &ValueDecl{}
	_ Node = &TypeDecl{}
	_ Node = &FileNode{}
	_ Node = &PackageNode{}
)

// ----------------------------------------
// Expr
//
// expressions generally have no side effects on the caller's context,
// except for channel blocks, type assertions, and panics.

type Expr interface {
	assertExpr()
	Node
}

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
func (*ConstExpr) assertExpr()        {}

var (
	_ Expr = &NameExpr{}
	_ Expr = &BasicLitExpr{}
	_ Expr = &BinaryExpr{}
	_ Expr = &CallExpr{}
	_ Expr = &IndexExpr{}
	_ Expr = &SelectorExpr{}
	_ Expr = &SliceExpr{}
	_ Expr = &StarExpr{}
	_ Expr = &RefExpr{}
	_ Expr = &TypeAssertExpr{}
	_ Expr = &UnaryExpr{}
	_ Expr = &CompositeLitExpr{}
	_ Expr = &KeyValueExpr{}
	_ Expr = &FuncLitExpr{}
	_ Expr = &ConstExpr{}
)

type Exprs []Expr

type NameExprType int

const (
	NameExprTypeNormal      NameExprType = iota // default
	NameExprTypeDefine                          // when defining normally
	NameExprTypeHeapDefine                      // when defining escaped name in loop
	NameExprTypeHeapUse                         // when above used in non-define lhs/rhs
	NameExprTypeHeapClosure                     // when closure captures name

	NameExprTypeLoopVarDefine // when defining a loopvar
	NameExprTypeLoopVarUse

	NameExprTypeLoopVarHeapDefine // when loopvar is captured
	NameExprTypeLoopVarHeapUse
)

type NameExpr struct {
	Attributes
	// TODO rename .Path's to .ValuePaths.
	Path ValuePath // set by preprocessor.
	Name
	Type NameExprType
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
	Func      Expr  // function expression
	Args      Exprs // function arguments, if any.
	Varg      bool  // if true, final arg is variadic.
	NumArgs   int   // len(Args) or len(Args[0].Results)
	WithCross bool  // if cross-called with `cur`.
}

// returns true if x is of form fn(cur,...) or fn(cross,...).
// but fn(cur,...) doesn't always mean with cross,
// because `cur` could be anything, so this is a sanity check.
func (x *CallExpr) isLikeWithCross() bool {
	if len(x.Args) == 0 {
		return false
	}
	first := x.Args[0]
	nx, ok := first.(*NameExpr)
	if !ok {
		return false
	}
	if nx.Name == Name("cross") || nx.Name == Name("cur") {
		return true
	}
	return false
}

// Legacy; only for fixing gno0.0 to gno0.9
func (x *CallExpr) isCrossing_gno0p0() bool {
	if x == nil {
		return false
	}
	if nx, ok := unconst(x.Func).(*NameExpr); ok {
		if nx.Name == "crossing" {
			return true
		}
	}
	return false
}

func (x *CallExpr) SetWithCross() {
	if !x.isLikeWithCross() {
		panic("expected fn(cur,...) or fn(cross,...)")
	}
	x.WithCross = true
}

func (x *CallExpr) IsWithCross() bool {
	return x.WithCross
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
func (x *CompositeLitExpr) IsKeyed() bool {
	return x.Elts.IsKeyed()
}

// A KeyValueExpr represents a single key-value pair in
// struct, array, slice, and map expressions.
type KeyValueExpr struct {
	Attributes
	Key   Expr // or nil
	Value Expr // never nil
}

type KeyValueExprs []KeyValueExpr

// Returns true if any elements are keyed.
// Panics if inconsistent.
func (kvxs KeyValueExprs) IsKeyed() bool {
	if len(kvxs) == 0 {
		return false
	} else if kvxs[0].Key == nil {
		for i := 1; i < len(kvxs); i++ {
			if kvxs[i].Key != nil {
				panic("mixed keyed and unkeyed elements")
			}
		}
		return false
	} else {
		for i := 1; i < len(kvxs); i++ {
			if kvxs[i].Key == nil {
				panic("mixed keyed and unkeyed elements")
			}
		}
		return true
	}
}

// A FuncLitExpr node represents a function literal.  Here one
// can reference statements from an expression, which
// completes the procedural circle.
type FuncLitExpr struct {
	Attributes
	StaticBlock
	Type         FuncTypeExpr // function type
	Body                      // function body
	HeapCaptures NameExprs    // filled in findLoopUses1
}

func (*FuncLitExpr) GetName() Name {
	return Name("")
}

func (fle *FuncLitExpr) GetFuncTypeExpr() *FuncTypeExpr {
	return &fle.Type
}

func (*FuncLitExpr) GetIsMethod() bool {
	return false
}

// The preprocessor replaces const expressions
// with *ConstExpr nodes.
type ConstExpr struct {
	Attributes
	Source Expr // (preprocessed) source of this value.
	TypedValue
	// Last BlockNode // consider (like constTypeExpr)
}

func NewConstExpr(source Expr, tv TypedValue) *ConstExpr {
	// internally, use toConstExpr().
	return toConstExpr(source, tv)
}

// ----------------------------------------
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
func (x *FieldTypeExpr) assertTypeExpr()     {}
func (x *ArrayTypeExpr) assertTypeExpr()     {}
func (x *SliceTypeExpr) assertTypeExpr()     {}
func (x *InterfaceTypeExpr) assertTypeExpr() {}
func (x *ChanTypeExpr) assertTypeExpr()      {}
func (x *FuncTypeExpr) assertTypeExpr()      {}
func (x *MapTypeExpr) assertTypeExpr()       {}
func (x *StructTypeExpr) assertTypeExpr()    {}
func (x *constTypeExpr) assertTypeExpr()     {}

func (x *FieldTypeExpr) assertExpr()     {}
func (x *ArrayTypeExpr) assertExpr()     {}
func (x *SliceTypeExpr) assertExpr()     {}
func (x *InterfaceTypeExpr) assertExpr() {}
func (x *ChanTypeExpr) assertExpr()      {}
func (x *FuncTypeExpr) assertExpr()      {}
func (x *MapTypeExpr) assertExpr()       {}
func (x *StructTypeExpr) assertExpr()    {}
func (x *constTypeExpr) assertExpr()     {}

var (
	_ TypeExpr = &FieldTypeExpr{}
	_ TypeExpr = &ArrayTypeExpr{}
	_ TypeExpr = &SliceTypeExpr{}
	_ TypeExpr = &InterfaceTypeExpr{}
	_ TypeExpr = &ChanTypeExpr{}
	_ TypeExpr = &FuncTypeExpr{}
	_ TypeExpr = &MapTypeExpr{}
	_ TypeExpr = &StructTypeExpr{}
	_ TypeExpr = &constTypeExpr{}
)

type FieldTypeExpr struct {
	Attributes
	NameExpr
	Type Expr

	// Currently only BasicLitExpr allowed.
	// NOTE: In Go, only struct fields can have tags.
	Tag Expr
}

type FieldTypeExprs []FieldTypeExpr

func (ftxz FieldTypeExprs) GetFieldTypeExpr(n Name) *FieldTypeExpr {
	for i := range ftxz {
		ftx := &ftxz[i]
		if ftx.Name == n {
			return ftx
		}
	}
	return nil
}

// Keep it slow, validating.
// If you need it faster, memoize it elsewhere.
func (ftxz FieldTypeExprs) IsNamed() bool {
	named := false
	for i, ftx := range ftxz {
		if i == 0 {
			if ftx.Name == "" || isMissingResult(ftx.Name) {
				named = false
			} else {
				named = true
			}
		} else {
			if named && (ftx.Name == "" || isMissingResult(ftx.Name)) {
				panic("[]FieldTypeExpr has inconsistent namedness (starts named)")
			} else if !named && (ftx.Name != "" && !isMissingResult(ftx.Name)) {
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

// Like ConstExpr but for types.
type constTypeExpr struct {
	Attributes
	Last   BlockNode // for GetTypeExprForExpr to resolve a *NameExpr.
	Source Expr      // (preprocessed) source of this value.
	Type   Type      // (jae) just `Type`? ConstExpr does it...
}

// ----------------------------------------
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

func (ss *Body) SetBody(nb Body) {
	*ss = nb
}

func (ss Body) GetLabeledStmt(label Name) (stmt Stmt, idx int) {
	for idx, stmt = range ss {
		if label == stmt.GetLabel() {
			return stmt, idx
		}
	}
	return nil, -1
}

// Legacy, only for fixing 0.0 to 0.9
func (ss Body) isCrossing_gno0p0() bool {
	if len(ss) == 0 {
		return false
	}
	fs := ss[0]
	xs, ok := fs.(*ExprStmt)
	if !ok {
		return false
	}
	cx, ok := xs.X.(*CallExpr)
	return ok && cx.isCrossing_gno0p0()
}

// ----------------------------------------

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
func (*RangeStmt) assertStmt()        {}
func (*ReturnStmt) assertStmt()       {}
func (*SelectStmt) assertStmt()       {}
func (*SelectCaseStmt) assertStmt()   {}
func (*SendStmt) assertStmt()         {}
func (*SwitchStmt) assertStmt()       {}
func (*SwitchClauseStmt) assertStmt() {}
func (*bodyStmt) assertStmt()         {}

var (
	_ Stmt = &AssignStmt{}
	_ Stmt = &BlockStmt{}
	_ Stmt = &BranchStmt{}
	_ Stmt = &DeclStmt{}
	_ Stmt = &DeferStmt{}
	_ Stmt = &EmptyStmt{}
	_ Stmt = &ExprStmt{}
	_ Stmt = &ForStmt{}
	_ Stmt = &GoStmt{}
	_ Stmt = &IfStmt{}
	_ Stmt = &IfCaseStmt{}
	_ Stmt = &IncDecStmt{}
	_ Stmt = &RangeStmt{}
	_ Stmt = &ReturnStmt{}
	_ Stmt = &SelectStmt{}
	_ Stmt = &SelectCaseStmt{}
	_ Stmt = &SendStmt{}
	_ Stmt = &SwitchStmt{}
	_ Stmt = &SwitchClauseStmt{}
	_ Stmt = &bodyStmt{}
)

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
	Op         Word  // keyword word (BREAK, CONTINUE, GOTO, FALLTHROUGH)
	Label      Name  // label name; or empty
	BlockDepth uint8 // blocks to pop
	FrameDepth uint8 // frames to pop
	BodyIndex  int   // index of statement of body
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
	Results     Exprs // result expressions; or nil
	CopyResults bool  // copy results to block first
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

// ----------------------------------------
// bodyStmt (persistent)

// NOTE: embedded in Block.
type bodyStmt struct {
	Attributes
	Body                       // for non-loop stmts
	BodyLen       int          // for for-continue
	NextBodyIndex int          // init:-2, cond/elem:-1, body:0..., post:n
	NumOps        int          // number of Ops, for goto
	NumValues     int          // number of Values, for goto
	NumExprs      int          // number of Exprs, for goto
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

func (x *bodyStmt) PopActiveStmt() (as Stmt) {
	as = x.Active
	x.Active = nil
	return
}

func (x *bodyStmt) LastStmt() Stmt {
	return x.Body[x.NextBodyIndex-1]
}

func (x *bodyStmt) String() string {
	next := ""
	if x.NextBodyIndex < 0 {
		next = "(init)"
	} else if x.NextBodyIndex == len(x.Body) {
		next = "(end)"
	} else {
		next = x.Body[x.NextBodyIndex].String()
	}
	active := ""
	if x.Active != nil {
		if x.NextBodyIndex < 0 || x.NextBodyIndex == len(x.Body) {
			// none
		} else if x.Body[x.NextBodyIndex-1] == x.Active {
			active = "*"
		} else {
			active = fmt.Sprintf(" unexpected active: %v", x.Active)
		}
	}
	return fmt.Sprintf("bodyStmt[%d/%d/%d]=%s%s Active:%v",
		x.ListLen,
		x.ListIndex,
		x.NextBodyIndex,
		next,
		active,
		x.Active)
}

// ----------------------------------------
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

// ----------------------------------------
// Decl

type Decl interface {
	Node
	GetDeclNames() []Name
	assertDecl()
}

type Decls []Decl

// non-pointer receiver to help make immutable.
func (x *FuncDecl) assertDecl()   {}
func (x *ImportDecl) assertDecl() {}
func (x *ValueDecl) assertDecl()  {}
func (x *TypeDecl) assertDecl()   {}

var (
	_ Decl = &FuncDecl{}
	_ Decl = &ImportDecl{}
	_ Decl = &ValueDecl{}
	_ Decl = &TypeDecl{}
)

// XXX consider embedding FuncLitExpr.
type FuncDecl struct {
	Attributes
	StaticBlock
	NameExpr
	IsMethod bool
	Recv     FieldTypeExpr // receiver (if method); or empty (if function)
	Type     FuncTypeExpr  // function signature: parameters and results
	Body                   // function body; or empty for external (non-Go) function

	unboundType *FuncTypeExpr // memoized
}

func (x *FuncDecl) GetName() Name {
	return x.NameExpr.Name
}

func (x *FuncDecl) GetDeclNames() []Name {
	if x.IsMethod {
		return nil
	} else {
		return []Name{x.NameExpr.Name}
	}
}

func (x *FuncDecl) GetFuncTypeExpr() *FuncTypeExpr {
	return &x.Type
}

func (x *FuncDecl) GetIsMethod() bool {
	return x.IsMethod
}

// *FuncDecl and *FuncLitExpr
type FuncNode interface {
	BlockNode
	GetName() Name // func lit expr returns ""
	GetFuncTypeExpr() *FuncTypeExpr
	GetIsMethod() bool
}

// If FuncDecl is for method, construct a FuncTypeExpr with receiver as first
// parameter.
func (x *FuncDecl) GetUnboundTypeExpr() *FuncTypeExpr {
	if x.IsMethod {
		if x.unboundType == nil {
			x.unboundType = &FuncTypeExpr{
				Attributes: x.Type.Attributes,
				Params:     append([]FieldTypeExpr{x.Recv}, x.Type.Params...),
				Results:    x.Type.Results,
			}
		}
		return x.unboundType
	}
	return &x.Type
}

type ImportDecl struct {
	Attributes
	NameExpr // local package name. required.
	PkgPath  string
}

func (x *ImportDecl) GetDeclNames() []Name {
	if x.NameExpr.Name == "." {
		return nil // ignore
	} else {
		return []Name{x.NameExpr.Name}
	}
}

type ValueDecl struct {
	Attributes
	NameExprs
	Type   Expr  // value type; or nil
	Values Exprs // initial value; or nil (unless const).
	Const  bool
}

func (x *ValueDecl) GetDeclNames() []Name {
	ns := make([]Name, 0, len(x.NameExprs))
	for _, nx := range x.NameExprs {
		if nx.Name == blankIdentifier {
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

func (x *TypeDecl) GetDeclNames() []Name {
	if x.NameExpr.Name == blankIdentifier {
		return nil // ignore
	} else {
		return []Name{x.NameExpr.Name}
	}
}

func HasDeclName(d Decl, n2 Name) bool {
	ns := d.GetDeclNames()
	return slices.Contains(ns, n2)
}

// ----------------------------------------
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

// ValueDecl and TypeDecl are the only decls that are both statements
// *and* decls.
func (x *ValueDecl) assertStmt() {}
func (x *TypeDecl) assertStmt()  {}

func (x *ValueDecl) assertSimpleDeclStmt() {}
func (x *TypeDecl) assertSimpleDeclStmt()  {}

var (
	_ SimpleDeclStmt = &ValueDecl{}
	_ SimpleDeclStmt = &TypeDecl{}
)

// ----------------------------------------
// *FileSet

type FileSet struct {
	Files []*FileNode
}

func (fs FileSet) GetFileNames() (fnames []string) {
	fnames = make([]string, 0, len(fs.Files))
	for _, fnode := range fs.Files {
		fnames = append(fnames, fnode.FileName)
	}
	return
}

func (fs *FileSet) AddFiles(fns ...*FileNode) {
	fs.Files = append(fs.Files, fns...)
}

func (fs *FileSet) GetFileByName(fname string) *FileNode {
	for _, fn := range fs.Files {
		if fn.FileName == fname {
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
	fn, decl, ok := fs.GetDeclForSafe(n)
	if !ok {
		panic(fmt.Sprintf(
			"name %s not defined in fileset with files %v",
			n, fs.FileNames()))
	}
	return fn, decl
}

func (fs *FileSet) GetDeclForSafe(n Name) (*FileNode, *Decl, bool) {
	// XXX index to bound to linear time.

	// Iteration happens reversing fs.Files; this is because the LAST declaration
	// of n is what we are looking for.
	for i := len(fs.Files) - 1; i >= 0; i-- {
		fn := fs.Files[i]
		for i, dn := range fn.Decls {
			if _, isImport := dn.(*ImportDecl); isImport {
				// imports in other files don't count.
				continue
			}
			if HasDeclName(dn, n) {
				// found the decl that declares n.
				return fn, &fn.Decls[i], true
			}
		}
	}
	return nil, nil, false
}

func (fs *FileSet) FileNames() []string {
	res := make([]string, len(fs.Files))
	for i, fn := range fs.Files {
		res[i] = fn.FileName
	}
	return res
}

// ----------------------------------------
// FileNode, & PackageNode

type FileNode struct {
	Attributes
	StaticBlock
	FileName string
	PkgName  Name
	Decls
}

type PackageNode struct {
	Attributes
	StaticBlock
	PkgPath  string
	PkgName  Name
	*FileSet // provides .GetDeclFor*()
}

func PackageNodeLocation(path string) Location {
	return Location{
		PkgPath: path,
	}
}

func NewPackageNode(name Name, path string, fset *FileSet) *PackageNode {
	pn := &PackageNode{
		PkgPath: path,
		PkgName: name,
		FileSet: fset,
	}
	pn.SetLocation(PackageNodeLocation(path))
	pn.InitStaticBlock(pn, nil)
	return pn
}

// PackageValue should be constructed here for initialization.
func (pn *PackageNode) NewPackage(alloc *Allocator) *PackageValue {
	var pv *PackageValue
	if pn.PkgName == "main" {
		// Allocation is only for the new created main package,
		// other packages are allocted while loading from store.
		pv = alloc.NewPackageValue(pn)
	} else {
		pv = &PackageValue{
			Block: &Block{
				Source: pn,
			},
			PkgName:    pn.PkgName,
			PkgPath:    pn.PkgPath,
			FNames:     nil,
			FBlocks:    nil,
			fBlocksMap: make(map[string]*Block),
		}
	}
	// Cannot set ObjectID here; it is not real yet.
	// BAD: pv.SetObjectID(ObjectIDFromPkgPath(pv.PkgPath))
	// Set realm for realm packages, main package, and ephemeral run packages
	if IsRealmPath(pn.PkgPath) || pn.PkgPath == "main" {
		rlm := NewRealm(pn.PkgPath)
		pv.SetRealm(rlm)
	} else if _, isRunPath := IsGnoRunPath(pn.PkgPath); isRunPath {
		rlm := NewRealm(pn.PkgPath)
		pv.SetRealm(rlm)
	}
	pv.IncRefCount() // all package values have starting ref count of 1.
	pn.PrepareNewValues(alloc, pv)
	return pv
}

// Prepares new func values (e.g. by attaching the proper file block closure).
// Returns a slice of new PackageValue.Values.
// After return, *PackageNode.Values and *PackageValue.Values have the same
// length. The implementation is similar to Block.ExpandWith.
// NOTE: declared methods do not get their closures set here. See
// *DeclaredType.GetValueAt() which returns a filled copy.
func (pn *PackageNode) PrepareNewValues(alloc *Allocator, pv *PackageValue) []TypedValue {
	// should already exist.
	block := pv.Block.(*Block)
	if block.Source != pn {
		// special case if block.Source is ref node
		if ref, ok := block.Source.(RefNode); ok && ref.Location == PackageNodeLocation(pv.PkgPath) {
			// this is fine
		} else {
			panic("PackageNode.PrepareNewValues() package mismatch")
		}
	}
	// The FuncValue Body may have been altered during the preprocessing.
	// We need to update body field from the source in the FuncValue accordingly.
	for _, tv := range pn.Values {
		if fv, ok := tv.V.(*FuncValue); ok {
			fv.UpdateBodyFromSource()
		}
	}
	pvl := len(block.Values)
	pnl := len(pn.Values)
	// copy new top-level defined values/types.
	if pvl < pnl {
		nvs := make([]TypedValue, pnl-pvl)
		copy(nvs, pn.Values[pvl:pnl])
		for i, tv := range nvs {
			if fv, ok := tv.V.(*FuncValue); ok {
				// copy function value and assign closure from package value.
				fv = fv.Copy(alloc) // mainly for main package func, MsgCall/MsgRun/filetest
				if fv.FileName == "" {
					// .uverse functions have no filename,
					// and repl runs declarations directly
					// on the package.
					fv.Parent = block
				} else {
					fb, ok := pv.fBlocksMap[fv.FileName]
					if !ok {
						// This is fine, it happens during pn.NewPackageValue()
						// panic(fmt.Sprintf("file block missing for file %q", fv.FileName))
					} else {
						fv.Parent = fb
					}
				}
				nvs[i].V = fv
			}
		}
		heapItems := pn.GetHeapItems()
		for i, tv := range nvs {
			if _, ok := tv.T.(heapItemType); ok {
				panic("unexpected heap item")
			}
			if heapItems[pvl+i] {
				nvs[i] = TypedValue{
					T: heapItemType{},
					V: alloc.NewHeapItem(nvs[i]),
				}
			}
		}
		block.Values = append(block.Values, nvs...)
		return block.Values[pvl:]
	} else if pvl > pnl {
		panic("package size error")
	} else {
		// nothing to do
		return nil
	}
}

// DefineNative defines a native function.
func (pn *PackageNode) DefineNative(n Name, ps, rs FieldTypeExprs, native func(*Machine)) {
	if debug {
		debug.Printf("*PackageNode.DefineNative(%s,...)\n", n)
	}
	if native == nil {
		panic("DefineNative expects a function, but got nil")
	}

	fd := FuncD(n, ps, rs, nil)
	fd = Preprocess(nil, pn, fd).(*FuncDecl)
	ft := evalStaticType(nil, pn, &fd.Type).(*FuncType)
	if debug {
		if ft == nil {
			panic("should not happen")
		}
	}
	fv := pn.GetSlot(nil, n, true).V.(*FuncValue)
	fv.nativeBody = native
}

// DefineNativeMethod defines a native method.
func (pn *PackageNode) DefineNativeMethod(r Name, n Name, ps, rs FieldTypeExprs, native func(*Machine)) {
	if debug {
		debug.Printf("*PackageNode.DefineNative(%s,...)\n", n)
	}
	if native == nil {
		panic("DefineNative expects a function, but got nil")
	}

	fd := MthdD(n, Fld("_", Nx(r)), ps, rs, nil)
	fd = Preprocess(nil, pn, fd).(*FuncDecl)
	ft := evalStaticType(nil, pn, &fd.Type).(*FuncType)
	if debug {
		if ft == nil {
			panic("should not happen")
		}
	}
	// attach fv to base declared type as method.
	nx := Preprocess(nil, pn, Nx(r)).(Expr)
	recv := evalStaticType(nil, pn, nx).(*DeclaredType)
	if debug {
		if ft == nil {
			panic("should not happen")
		}
	}
	// recv.DefineMethod(fv)
	path := recv.GetPathForName(n)
	fv := recv.Methods[path.Index].GetFunc()
	fv.nativeBody = native
}

// Same as DefineNative but allow the overriding of previously defined natives.
// For example, overriding a native function defined in stdlibs/stdlibs for
// testing. Caller must ensure that the function type is identical.
func (pn *PackageNode) DefineNativeOverride(n Name, native func(*Machine)) {
	if debug {
		debug.Printf("*PackageNode.DefineNativeOverride(%s,...)\n", n)
	}
	if native == nil {
		panic("DefineNative expects a function, but got nil")
	}
	fv := pn.GetSlot(nil, n, true).V.(*FuncValue)
	fv.nativeBody = native
}

// ----------------------------------------
// RefNode

// Reference to a node by its location.
type RefNode struct {
	Location  // location of node.
	BlockNode // convenience to implement BlockNode (nil).
}

func (ref RefNode) GetLocation() Location {
	return ref.Location.GetLocation()
}

func (ref RefNode) SetLocation(loc Location) {
	// NOTE: Keep RefNode a non-pointer type,
	// and disallow RefNode.SetLocation().
	// You can still call ref.Location.SetLocation().
	panic("should not happen")
}

// ----------------------------------------
// BlockNode

// Nodes that create their own scope satisfy this interface.
type BlockNode interface {
	Node
	InitStaticBlock(BlockNode, BlockNode)
	IsInitialized() bool
	GetStaticBlock() *StaticBlock
	GetLocation() Location
	SetLocation(Location)

	// StaticBlock promoted methods
	GetParentNode(Store) BlockNode
	Reserve(bool, *NameExpr, Node, NSType, int)
	Define(Name, TypedValue)
	Define2(bool, Name, Type, TypedValue, NameSource)
	GetPathForName(Store, Name) ValuePath
	GetBlockNames() []Name
	GetExternNames() []Name
	GetNumNames() uint16
	GetIsConst(Store, Name) bool
	GetIsConstAt(Store, ValuePath) bool
	GetLocalIndex(Name) (uint16, bool)
	GetSlot(Store, Name, bool) *TypedValue // was GetValueRef()
	SetIsHeapItem(n Name)
	GetHeapItems() []bool
	GetBlockNodeForPath(Store, ValuePath) BlockNode
	GetStaticTypeOf(Store, Name) Type
	GetStaticTypeOfAt(Store, ValuePath) Type
	GetBody() Body
	SetBody(Body)

	FindNameMaybeLoopvar(Store, Name) (bool, bool)

	// Utility methods for gno fix etc.
	// Unlike GetType[Decl|Expr]For[Path|Expr] which are determined
	// statically, functions may be variable, so GetFuncNodeFor[Path|Expr]
	// may return an error if the func node cannot be determined.
	// (vs say GetTypeDeclForPath() is user error if it panics).
	GetNameSources() []NameSource
	GetNameSourceForPath(Store, ValuePath) (BlockNode, *FileNode, NameSource)
	GetTypeDeclForPath(Store, ValuePath) *TypeDecl
	GetTypeDeclForExpr(Store, Expr) *TypeDecl
	GetTypeExprForPath(Store, ValuePath) (BlockNode, TypeExpr)
	GetTypeExprForExpr(Store, Expr) (BlockNode, TypeExpr)
	GetFuncNodeForPath(Store, ValuePath) (FuncNode, error)
	GetFuncNodeForExpr(Store, Expr) (FuncNode, error)
}

var (
	_ BlockNode = &FuncLitExpr{}
	_ BlockNode = &BlockStmt{}
	_ BlockNode = &ForStmt{}
	_ BlockNode = &IfStmt{} // faux block node
	_ BlockNode = &IfCaseStmt{}
	_ BlockNode = &RangeStmt{}
	_ BlockNode = &SelectCaseStmt{}
	_ BlockNode = &SwitchStmt{} // faux block node
	_ BlockNode = &SwitchClauseStmt{}
	_ BlockNode = &FuncDecl{}
	_ BlockNode = &FileNode{}
	_ BlockNode = &PackageNode{}
	_ BlockNode = RefNode{}
)

// ----------------------------------------
// StaticBlock

// Embed in node to make it a BlockNode.
type StaticBlock struct {
	Block
	Location
	Types             []Type
	NumNames          uint16
	Names             []Name
	NameSources       []NameSource
	HeapItems         []bool
	UnassignableNames []Name
	Consts            []Name // TODO consider merging with Names.
	Externs           []Name
	Parent            BlockNode

	// temporary storage for rolling back redefinitions.
	oldValues []oldValue
}

// NameSource holds origin information about a name.
type NameSource struct {
	NameExpr *NameExpr // name expr of decl/assign/etc
	Origin   Node      // ref to a node in block.Source
	Type     NSType    // type of name
	Index    int       // index given type, or -1
}

var noNameSource = NameSource{}

func (nsrc NameSource) IsZero() bool {
	return nsrc == noNameSource
}

type NSType int // name source type

const (
	NSDefine NSType = iota // AssignStmt <name>... := (indexed)
	NSImportDecl
	NSValueDecl    // var <name>... (indexed)
	NSTypeDecl     // type <name>
	NSFuncDecl     // func <name>
	NSRangeKey     // for <name> := range
	NSRangeValue   // for _, <name> := range
	NSFuncReceiver // func (<name> _) _()
	NSFuncParam    // func(<name>...) (indexed)
	NSFuncResult   // func()<name>... (indexed)
	NSTypeSwitch   // switch <name> := _.(type)
)

type oldValue struct {
	idx   uint16
	value Value
}

// revert values upon failure of redefinitions.
func (sb *StaticBlock) revertToOld() {
	for _, ov := range sb.oldValues {
		sb.Block.Values[ov.idx].V = ov.value
	}
	sb.oldValues = nil
}

// Implements BlockNode
func (sb *StaticBlock) InitStaticBlock(source BlockNode, parent BlockNode) {
	if sb.Names != nil || sb.Block.Source != nil {
		panic("StaticBlock already initialized")
	}
	if parent == nil {
		sb.Block = Block{
			Source: source,
			Values: nil,
			Parent: nil,
		}
	} else {
		switch source.(type) {
		case *IfCaseStmt, *SwitchClauseStmt:
			if parent == nil {
				sb.Block = Block{
					Source: source,
					Values: nil,
					Parent: nil,
				}
			} else {
				parent2 := parent.GetParentNode(nil)
				sb.Block = Block{
					Source: source,
					Values: nil,
					Parent: parent2.GetStaticBlock().GetBlock(),
				}
			}
		default:
			sb.Block = Block{
				Source: source,
				Values: nil,
				Parent: parent.GetStaticBlock().GetBlock(),
			}
		}
	}
	sb.NumNames = 0
	sb.Names = make([]Name, 0, 16)
	sb.NameSources = make([]NameSource, 0, 16)
	sb.HeapItems = make([]bool, 0, 16)
	sb.Consts = make([]Name, 0, 16)
	sb.Externs = make([]Name, 0, 16)
	sb.Parent = parent
}

// Implements BlockNode.
func (sb *StaticBlock) IsInitialized() bool {
	return sb.Block.Source != nil
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
	return sb.Names
}

// Implements BlockNode.
// NOTE: Extern names may also be local, if declared after usage as an extern
// (thus shadowing the extern name).
func (sb *StaticBlock) GetExternNames() (ns []Name) {
	return sb.Externs
}

func (sb *StaticBlock) addExternName(n Name) {
	if slices.Contains(sb.Externs, n) {
		return
	}
	sb.Externs = append(sb.Externs, n)
}

// Implements BlockNode.
func (sb *StaticBlock) GetNumNames() (nn uint16) {
	return sb.NumNames
}

// Implements BlockNode.
func (sb *StaticBlock) GetHeapItems() []bool {
	return sb.HeapItems
}

// Implements BlockNode.
func (sb *StaticBlock) SetIsHeapItem(n Name) {
	idx, ok := sb.GetLocalIndex(n)
	if !ok {
		panic("name not found in block")
	}
	sb.HeapItems[idx] = true
}

// Implements BlockNode.
func (sb *StaticBlock) GetParentNode(store Store) BlockNode {
	return sb.Parent
}

// Implements BlockNode.
// As a side effect, notes externally defined names.
// Slow, for precompile only.
func (sb *StaticBlock) GetPathForName(store Store, n Name) ValuePath {
	if n == blankIdentifier {
		return NewValuePathBlock(0, 0, blankIdentifier)
	}
	// Check local.
	gen := 1
	if idx, ok := sb.GetLocalIndex(n); ok {
		return NewValuePathBlock(uint8(gen), idx, n)
	}
	sn := sb.GetSource(store)
	// Register as extern.
	// NOTE: uverse names are externs too.
	// NOTE: externs may also be shadowed later in the block. Thus, usages
	// before the declaration will have depth > 1; following it, depth == 1,
	// matching the two different identifiers they refer to.
	if !isFile(sn) {
		sb.GetStaticBlock().addExternName(n)
	}
	// Check ancestors.
	gen++
	fauxChild := 0
	if fauxChildBlockNode(sn) {
		fauxChild++
	}
	sn = sn.GetParentNode(store)
	for sn != nil {
		if idx, ok := sn.GetLocalIndex(n); ok {
			if 0xff < (gen - fauxChild) {
				panic("value path depth overflow")
			}
			return NewValuePathBlock(uint8(gen-fauxChild), idx, n)
		} else {
			if !isFile(sn) {
				sn.GetStaticBlock().addExternName(n)
			}
			gen++
			if fauxChildBlockNode(sn) {
				fauxChild++
			}
			sn = sn.GetParentNode(store)
		}
	}
	// Finally, check uverse.
	if idx, ok := UverseNode().GetLocalIndex(n); ok {
		return NewValuePathUverse(idx, n)
	}
	// Name does not exist.
	panic(fmt.Sprintf("name %s not declared", n))
}

// Get the containing block node for node with path relative to this containing block.
// Slow, for precompile only.
func (sb *StaticBlock) GetBlockNodeForPath(store Store, path ValuePath) BlockNode {
	if path.Type == VPUverse {
		return UverseNode()
	}
	if path.Type != VPBlock {
		panic("expected block type value path but got " + path.Type.String())
	}

	// NOTE: path.Depth == 1 means it's in bn.
	bn := sb.GetSource(store)

	for i := 1; i < int(path.Depth); i++ {
		if fauxChildBlockNode(bn) {
			bn = bn.GetParentNode(store)
		}
		bn = bn.GetParentNode(store)
	}

	// If bn is a faux child block node, check also its faux parent.
	switch bn := bn.(type) {
	case *IfCaseStmt, *SwitchClauseStmt:
		pn := bn.GetParentNode(store)
		if path.Index < pn.GetNumNames() {
			return pn
		}
	}

	return bn
}

// Returns whether a name defined here in in ancestry is a const.
// This is not the same as whether a name's static type is
// untyped -- as in c := a == b, a name may be an untyped non-const.
// Implements BlockNode.
func (sb *StaticBlock) GetIsConst(store Store, n Name) bool {
	_, ok := sb.GetLocalIndex(n)
	bp := sb.GetParentNode(store)
	for {
		if ok {
			return sb.getLocalIsConst(n)
		} else if bp != nil {
			_, ok = bp.GetLocalIndex(n)
			sb = bp.GetStaticBlock()
			bp = bp.GetParentNode(store)
		} else {
			panic(fmt.Sprintf("name %s not declared", n))
		}
	}
}

func (sb *StaticBlock) GetIsConstAt(store Store, path ValuePath) bool {
	return sb.GetBlockNodeForPath(store, path).GetStaticBlock().getLocalIsConst(path.Name)
}

// Returns true iff n is a local const defined name.
func (sb *StaticBlock) getLocalIsConst(n Name) bool {
	return slices.Contains(sb.Consts, n)
}

func (sb *StaticBlock) IsAssignable(store Store, n Name) bool {
	_, ok := sb.GetLocalIndex(n)
	bp := sb.GetParentNode(store)
	un := sb.UnassignableNames

	for {
		if ok {
			return !slices.Contains(un, n)
		} else if bp != nil {
			_, ok = bp.GetLocalIndex(n)
			un = bp.GetStaticBlock().UnassignableNames
			bp = bp.GetParentNode(store)
		} else if _, ok := UverseNode().GetLocalIndex(n); ok {
			return false
		} else {
			return true
		}
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
		} else if idx, ok := UverseNode().GetLocalIndex(n); ok {
			path := NewValuePathUverse(idx, n)
			tv := Uverse().GetValueAt(store, path)
			return tv.T
		} else {
			panic(fmt.Sprintf("name %s not declared", n))
		}
	}
}

// Implements BlockNode.
func (sb *StaticBlock) GetStaticTypeOfAt(store Store, path ValuePath) Type {
	if debug {
		if path.Depth == 0 {
			panic("should not happen")
		}
	}
	bn := sb.GetBlockNodeForPath(store, path)
	return bn.GetStaticBlock().Types[path.Index]
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

func (sb *StaticBlock) FindNameMaybeLoopvar(store Store, n Name) (loopvar, found bool) {
	fmt.Println("FindNameSkipPredefined, n: ", n)
	if n == blankIdentifier {
		return false, false
	}
	// Check local.
	gen := 1
	// also search with .loopvar_, this make sure `i` also
	// get a correct path.
	if _, loopvar, found = sb.GetLocalIndexMaybeLoopvar(n); found {
		fmt.Println("===loopVar: ", loopvar)
		// found a NameExpr with type NameExprTypeLoopVarDefine
		return
	}
	// Check ancestors.
	gen++
	bp := sb.GetParentNode(store)
	for bp != nil {
		if _, loopvar, found = bp.GetStaticBlock().GetLocalIndexMaybeLoopvar(n); found {
			// found a NameExpr with type NameExprTypeLoopVarDefine
			return loopvar, found
		} else {
			bp = bp.GetParentNode(store)
			gen++
			if 0xff < gen {
				panic("value path depth overflow")
			}
		}
	}
	return
}

func (sb *StaticBlock) GetLocalIndexMaybeLoopvar(n Name) (uint16, bool, bool) {
	// fmt.Println("===GetLocalIndexSkipPredefined, sb: ", sb.Block)
	// fmt.Println("===GetLocalIndexSkipPredefined, n: ", n)
	// if loopvar is found.
	var loopvar bool

	// firstly search general TypeDefine names,
	// it potentially overrides the loopvar.
	for i, name := range sb.Names {
		if name == n {
			if debug {
				nt := reflect.TypeOf(sb.Source).String()
				debug.Printf("StaticBlock(%p %v).GetLocalIndex(%s) = %v, %v\n",
					sb, nt, n, i, name)
			}
			// skip predefined name
			t := sb.Types[i]
			if t != nil {
				return uint16(i), loopvar, true
			}
			// else going on search loopvar
		}
	}

	// if not found above, looking for loopvar.
	n2 := Name(fmt.Sprintf(".loopvar_%s", n))
	// fmt.Println("===n2: ", n2)
	for i, name := range sb.Names {
		// println("===search loopvar")
		if name == n2 {
			if debug {
				nt := reflect.TypeOf(sb.Source).String()
				debug.Printf("StaticBlock(%p %v).GetLocalIndex(%s) = %v, %v\n",
					sb, nt, n, i, name)
			}

			loopvar = true

			// XXX, skip predefine name, why?
			t := sb.Types[i]
			if t == nil {
				return 0, loopvar, false
			}
			return uint16(i), loopvar, true
		}
	}
	if debug {
		nt := reflect.TypeOf(sb.Source).String()
		debug.Printf("StaticBlock(%p %v).GetLocalIndex(%s) = undefined\n",
			sb, nt, n)
	}
	return 0, loopvar, false
}

func processLoopVar(last BlockNode, nx *NameExpr) {
	// fmt.Println("===renameLoopVar, nx: ", nx, nx.Type)
	if nx.Name == blankIdentifier {
		return
	}

	if nx.Type == NameExprTypeNormal {
		// handle loopvar stuff
		loopvar, found := last.FindNameMaybeLoopvar(nil, nx.Name)
		if found && loopvar {
			fmt.Println("---found loopvar use, nx: ", nx)
			nx.Type = NameExprTypeLoopVarUse
			// XXX, necessary?
			nx.Name = Name(fmt.Sprintf(".loopvar_%s", nx.Name))
			fmt.Println("===after rename, nx: ", nx)
		} else {
			fmt.Println("Not loopvar, nx: ", nx, nx.Type)
		}
	}
}

// Implemented BlockNode.
// This method is too slow for runtime, but it is used during preprocessing to
// compute types.  If ignoreReserved, skips over names that are only reserved
// (and neither predefined nor defined).  Returns nil if not found.
func (sb *StaticBlock) GetSlot(store Store, n Name, ignoreReserved bool) *TypedValue {
	idx, ok := sb.GetLocalIndex(n)
	bb := &sb.Block
	bp := sb.GetParentNode(store)
	for {
		if ok && (!ignoreReserved || sb.Types[idx] != nil) {
			return bb.GetPointerToInt(store, int(idx)).TV
		} else if bp != nil {
			idx, ok = bp.GetLocalIndex(n)
			sb = bp.GetStaticBlock()
			bb = sb.GetBlock()
			bp = bp.GetParentNode(store)
		} else {
			return nil
		}
	}
}

// Implements BlockNode.
// NOTE: This probably isn't what you think it is.
// See GetNameSourceForPath() impl.
func (sb *StaticBlock) GetNameSources() []NameSource {
	return sb.NameSources
}

// Implemented BlockNode.
// Convenience for getting name origin and source name expr.
// Too slow for runtime.
// The returned *NameExpr is in the context of the filenode if BlockNode is a
// package node, otherwise is in the block node.  See also usage of `skipFile`.
// NOTE The returned *NameExpr is used by `gno fix` to store attributes.
func (sb *StaticBlock) GetNameSourceForPath(store Store, path ValuePath) (BlockNode, *FileNode, NameSource) {
	dbn := sb.GetBlockNodeForPath(store, path)
	nsrc := dbn.GetNameSources()[path.Index]
	var fn *FileNode
	if pn, ok := dbn.(*PackageNode); ok {
		fn, _ = pn.GetDeclFor(nsrc.NameExpr.Name)
	} else {
		fname := dbn.GetLocation().GetFile()
		pn := packageOf(dbn)
		fn = pn.GetFileByName(fname)
	}
	return dbn, fn, nsrc
}

// Implemented BlockNode.
// This method is too slow for runtime, but it is used by `gno fix` to find the
// origin type declaration.  Panics if path does not lead to a type decl.
//
// NOTE: Types are interchangeable, or should be, so they should not be used
// for acquiring the source in general, unless it is a struct or interface type
// which may have unexposed names; and for these they also need to keep
// .PkgPath; but still should not be relied on for acquiring source.
// Use GetType[Decl|Expr]For[Expr|Path]() instead.
func (sb *StaticBlock) GetTypeDeclForPath(store Store, path ValuePath) *TypeDecl {
	if path.Type != VPBlock {
		panic(fmt.Sprintf("expected path.Type of VPBlock but got %v", path))
	}
	dbn := sb.GetBlockNodeForPath(store, path)
	td := dbn.GetNameSources()[path.Index].Origin.(*TypeDecl)
	return td
}

// Implemented BlockNode.
// This method is too slow for runtime, but it is used by `gno fix` to find the
// origin type declaration.  Panics if type does not lead to a type expr, but
// if the type was elided returns nil.  Valid types are those that can be .Type
// in a TypeDecl; *TypeExpr, *NameExpr, and *SelectorExpr.
//
// See also note for GetTypeDeclForPath.
func (sb *StaticBlock) GetTypeDeclForExpr(store Store, txe Expr) *TypeDecl {
	switch txe := txe.(type) {
	case *constTypeExpr:
		return txe.Last.GetTypeDeclForExpr(store, txe.Source)
	case *TypeDecl:
		return txe
	case TypeExpr:
		panic(fmt.Sprintf("unexpected type expr %v (a type expr cannot refer to a type decl)", txe))
	case *NameExpr:
		if txe.Name == ".elided" {
			// It's not in general easy or fast to find the TypeDecl from
			// elided types if the type is unnamed, because the store
			// doesn't store type exprs by location (they don't have one).
			// We could store the parent loc in every Type and perform a
			// search, but for now we return nil instead of panic'ing.  See
			// *NameExpr ".elided" below.
			return nil
		}
		// With type-aliases to an external type like `type Foo =
		// extrealm.Bar`, the *NameExpr doesn't correspond to sb.
		return sb.GetTypeDeclForPath(store, txe.Path)
	case *SelectorExpr:
		switch txex := txe.X.(type) {
		case *ConstExpr:
			pn := txex.V.(*PackageNode)
			return pn.GetTypeDeclForPath(store, txe.Path)
		case *NameExpr:
			pkgPath := txex.GetAttribute(ATTR_PACKAGE_PATH)
			if pkgPath == nil {
				panic(fmt.Sprintf("unexpected name expr %v that isn't a package", txex))
			}
			pn := store.GetPackageNode(pkgPath.(string))
			return pn.GetTypeDeclForPath(store, txe.Path)
		default:
			panic(fmt.Sprintf("unexpected expr %v", txex))
		}
	default:
		panic(fmt.Sprintf("expected expr (to refer to a type decl) but got %v (%T)", txe, txe))
	}
}

// Implemented BlockNode.
// This method is too slow for runtime, but it is used by `gno fix` to find the
// origin type declaration.  Unlike GetTypeDeclForPath, this function is
// recursive because a type decl may refer to a type-expr by a non-type-expr
// expression such as a name expr or selector.  The returned block node is the
// one where the type expr is actually declared, typically the file node.
// Panics if path does not lead to a type expr.
//
// See also note for GetTypeDeclForPath.
func (sb *StaticBlock) GetTypeExprForPath(store Store, path ValuePath) (BlockNode, TypeExpr) {
	if path.Type != VPBlock {
		panic(fmt.Sprintf("expected path.Type of VPBlock but got %v", path))
	}
	td := sb.GetTypeDeclForPath(store, path)
	return sb.GetTypeExprForExpr(store, td.Type)
}

// Implemented BlockNode.
// This method is too slow for runtime, but it is used by `gno fix` to find the
// origin type declaration.  Valid types are those that can be .Type in a
// TypeDecl; *TypeExpr, *NameExpr, and *SelectorExpr.  The returned block node
// is the one where the type expr is actually declared, typically the file
// node.  If txe is a type expr, returns the source for sb. Panics if type does
// not lead to a type expr.
//
// See also note for GetTypeDeclForPath.
func (sb *StaticBlock) GetTypeExprForExpr(store Store, txe Expr) (BlockNode, TypeExpr) {
	switch txe := txe.(type) {
	case *constTypeExpr:
		// NOTE: `last` usually refers to a file node for file decls.
		return txe.Last.GetTypeExprForExpr(store, txe.Source)
	case *TypeDecl:
		return sb.GetTypeExprForExpr(store, txe.Type)
	case TypeExpr:
		return sb.Block.GetSource(store), txe
	case *NameExpr:
		if txe.Name == ".elided" {
			// It's not in general easy or fast to find the TypeExpr from
			// elided types if the type is unnamed, because the store
			// doesn't store type exprs by location (they don't have one).
			// We could store the parent loc in every Type and perform a
			// search, but for now we return nil instead of panic'ing.  See
			// *NameExpr ".elided" below.
			return nil, nil
		}
		// With type-aliases to an external type like `type Foo =
		// extrealm.Bar`, the *NameExpr doesn't correspond to sb.
		return sb.GetTypeExprForPath(store, txe.Path)
	case *SelectorExpr:
		switch txex := txe.X.(type) {
		case *ConstExpr:
			pn := txex.V.(*PackageNode)
			return pn.GetTypeExprForPath(store, txe.Path)
		case *NameExpr:
			pkgPath := txex.GetAttribute(ATTR_PACKAGE_PATH)
			if pkgPath == nil {
				panic(fmt.Sprintf("unexpected name expr %v that isn't a package", txex))
			}
			pn := store.GetPackageNode(pkgPath.(string))
			return pn.GetTypeExprForPath(store, txe.Path)
		default:
			panic(fmt.Sprintf("unexpected expr %v", txex))
		}
	default:
		panic(fmt.Sprintf("expected type expr (suitable for type decl) but got %v (%T)", txe, txe))
	}
}

// Implemented BlockNode.
// This method is too slow for runtime, but it is used by `gno fix` to find the
// origin func decl/expr.
func (sb *StaticBlock) GetFuncNodeForPath(store Store, path ValuePath) (FuncNode, error) {
	if path.Type != VPBlock && path.Type != VPUverse {
		return nil, fmt.Errorf("expected path.Type of VPBlock but got %v", path)
	}
	dbn := sb.GetBlockNodeForPath(store, path)
	fn := dbn.GetNameSources()[path.Index].Origin
	switch fn := fn.(type) {
	case *FuncDecl:
		return fn, nil
	case *FuncLitExpr:
		return fn, nil
	default:
		return nil, fmt.Errorf("unexpected node type %v (expected func node)", fn)
	}
}

// Implemented BlockNode.
// This method is too slow for runtime, but it is used by `gno fix` to find the
// origin func decl/expr.
func (sb *StaticBlock) GetFuncNodeForExpr(store Store, fne Expr) (FuncNode, error) {
	fne = unconst(fne)
	switch fne := fne.(type) {
	case FuncNode:
		return fne, nil
	case *NameExpr:
		return sb.GetFuncNodeForPath(store, fne.Path)
	case *SelectorExpr:
		fnsx, ok := fne.X.(*ConstExpr)
		if !ok {
			return nil, fmt.Errorf("unhandled selector base in %v (expected something const)", fne)
		}
		switch fnsxt := fnsx.T.(type) {
		case *PackageType:
			var pn *PackageNode
			pv, ok := fnsx.V.(*PackageValue)
			if !ok {
				ref, ok := fnsx.V.(RefValue)
				if !ok {
					// this shouldn't happen, thus a panic.
					panic(fmt.Sprintf("unexpected package value type %T", fnsx.V))
				}
				this := packageOf(sb.GetSource(store))
				if ref.PkgPath == this.PkgPath {
					// NOTE: when non-integration *_test.gno refer
					// to a selector(ref(pkgPath),<decl name>), the
					// selector's path will not match because the
					// import from store doesn't include decls from
					// *_test.gno, and also the files are sorted
					// alphabetically so path indices will be off.
					// So just return this.
					pn = this
				} else {
					pn = store.GetPackageNode(ref.PkgPath)
				}
			} else {
				pn = pv.GetBlock(store).GetSource(store).(*PackageNode)
				// pn = store.GetPackageNode(ref.PkgPath)
			}
			return pn.GetFuncNodeForPath(store, fne.Path)
		case *DeclaredType:
			switch fne.Path.Type {
			case VPValMethod, VPPtrMethod:
				mtv := fnsxt.Methods[fne.Path.Index]
				return mtv.V.(*FuncValue).GetSource(store).(FuncNode), nil
			default:
				return nil, fmt.Errorf("%v is not a method function", fne)
			}
		default:
			return nil, fmt.Errorf("unhandled selector base in %v (expected package or decl)", fne)
		}
	default:
		return nil, fmt.Errorf("unhandled expr (expected func node, name expr, or selector expr) but "+
			"got %v", fne)
	}
}

// Implements BlockNode
// Statically declares a name definition.
// At runtime, use *Block.GetPointerTo() which takes a path
// value, which is pre-computeed in the preprocessor.
// Once a typed value is defined, it cannot be changed.
//
// NOTE: Currently tv.V is only set when the value represents a Type(Value) or
// a FuncValue.  The purpose of tv is to describe the invariant of a named
// value, at the minimum its type, but also sometimes the typeval value; but we
// could go further and store preprocessed constant results here too.  See
// "anyValue()" and "asValue()" for usage.
func (sb *StaticBlock) Define(n Name, tv TypedValue) {
	sb.Define2(false, n, tv.T, tv, NameSource{})
}

// Set type to nil, only reserving the name.
func (sb *StaticBlock) Reserve(isConst bool, nx *NameExpr, origin Node, nstype NSType, index int) {
	_, exists := sb.GetLocalIndex(nx.Name)
	if !exists {
		sb.Define2(isConst, nx.Name, nil, anyValue(nil), NameSource{nx, origin, nstype, index})
	}
}

// The declared type st may not be the same as the static tv;
// e.g. var x MyInterface = MyStruct{}.
// Setting st and tv to nil/zero reserves (predefines) name for definition later.
func (sb *StaticBlock) Define2(isConst bool, n Name, st Type, tv TypedValue, nsrc NameSource) {
	if debug {
		debug.Printf(
			"StaticBlock.Define2(%v, %s, %v, %v)\n", // XXX add nsrc
			isConst, n, st, tv)
	}
	// TODO check that tv.T implements t.
	if len(n) == 0 {
		panic("name cannot be zero")
	}
	if int(sb.NumNames) != len(sb.Names) {
		panic("StaticBlock.NumNames and len(.Names) mismatch")
	}
	if int(sb.NumNames) != len(sb.Types) {
		panic("StaticBlock.NumNames and len(.Types) mismatch")
	}
	if int(sb.NumNames) != len(sb.NameSources) {
		panic("StaticBlock.NumNames and len(.NameSources) mismatch")
	}
	if sb.NumNames == math.MaxUint16 {
		panic("too many variables in block")
	}
	if tv.T == nil && tv.V != nil {
		panic("StaticBlock.Define2() requires .T if .V is set")
	}
	if n == blankIdentifier {
		return // ignore
	}
	idx, exists := sb.GetLocalIndex(n)
	if exists {
		// Is re-defining.
		if isConst != sb.getLocalIsConst(n) {
			panic(fmt.Sprintf(
				"StaticBlock.Define2(%s) cannot change const status",
				n))
		}
		old := sb.Block.Values[idx]
		if !old.IsUndefined() && tv.T != nil {
			if tv.T.Kind() == FuncKind && baseOf(tv.T).(*FuncType).IsZero() {
				// special case,
				// allow re-predefining for func upgrades.
				// keep the old type so we can check it at preprocessor.
				tv.T = old.T
				fv := tv.V.(*FuncValue)
				fv.Type = old.T
				st = old.T
				sb.oldValues = append(sb.oldValues,
					oldValue{idx, old.V})
			} else {
				if tv.T.TypeID() != old.T.TypeID() {
					panic(fmt.Sprintf(
						"StaticBlock.Define2(%s) cannot change .T; was %v, new %v",
						n, old.T, tv.T))
				}
				if tv.V != old.V {
					panic(fmt.Sprintf(
						"StaticBlock.Define2(%s) cannot change .V",
						n))
				}
			}
			// Allow re-definitions if they have the same type.
			// (In normal scenarios, duplicate declarations are "caught" by RunMemPackage.)
		}
		sb.Block.Values[idx] = tv
		sb.Types[idx] = st
		if !nsrc.IsZero() {
			sb.NameSources[idx] = nsrc
		}
		// This can happen when a *ValueDecl is split after initStaticBlocks.
		/*
			if !nsrc.IsZero() && sb.NameSources[idx] != nsrc {
				panic(fmt.Sprintf("name source mismatch in block Define2(): was %v, got %v",
					sb.NameSources[idx], nsrc))
			}
		*/
	} else {
		// The general case without re-definition.
		sb.Names = append(sb.Names, n)
		sb.HeapItems = append(sb.HeapItems, false)
		if isConst {
			sb.Consts = append(sb.Consts, n)
		}
		sb.NumNames++
		sb.Block.Values = append(sb.Block.Values, tv)
		sb.Types = append(sb.Types, st)
		sb.NameSources = append(sb.NameSources, nsrc)
	}
}

// Implements BlockNode
func (sb *StaticBlock) SetStaticBlock(osb StaticBlock) {
	*sb = osb
}

func (x *IfStmt) GetBody() Body {
	panic("IfStmt has no body (but .Then and .Else do)")
}

func (x *IfStmt) SetBody(b Body) {
	panic("IfStmt has no body (but .Then and .Else do)")
}

func (x *SwitchStmt) GetBody() Body {
	panic("SwitchStmt has no body (but its cases do)")
}

func (x *SwitchStmt) SetBody(b Body) {
	panic("SwitchStmt has no body (but its cases do)")
}

func (x *FileNode) GetBody() Body {
	panic("FileNode has no body (but it does have .Decls)")
}

func (x *FileNode) SetBody(b Body) {
	panic("FileNode has no body (but it does have .Decls)")
}

func (*PackageNode) GetBody() Body {
	panic("PackageNode has no body")
}

func (*PackageNode) SetBody(b Body) {
	panic("PackageNode has no body")
}

// ----------------------------------------
// Value Path

// A relative pointer to a TypedValue value
//
//	(a) a Block scope var or const
//	(b) a StructValue field
//	(c) a DeclaredType method
//	(d) a PackageNode declaration
//
// Depth tells how many layers of access should be unvealed before
// arriving at the ultimate handler type.  In the case of Blocks,
// the depth tells how many layers of ancestry to ascend before
// arriving at the target block.  For other selector expr paths
// such as those for *DeclaredType methods or *StructType fields,
// see tests/selector_test.go.
type ValuePath struct {
	Type VPType // see VPType* consts.
	// Warning: Use SetDepth() to set Depth.
	Depth uint8  // see doc for ValuePath.
	Index uint16 // index of value, field, or method.
	Name  Name   // name of value, field, or method.
}

// Maximum depth of a ValuePath.
const MaxValuePathDepth = 127

func (vp ValuePath) validateDepth() {
	if vp.Depth > MaxValuePathDepth {
		panic(fmt.Sprintf("exceeded maximum %s depth (%d)", vp.Type, MaxValuePathDepth))
	}
}

func (vp *ValuePath) SetDepth(d uint8) {
	vp.Depth = d

	vp.validateDepth()
}

type VPType uint8

const (
	VPInvalid        VPType = 0x00 // not used
	VPUverse         VPType = 0x01
	VPBlock          VPType = 0x02 // blocks and packages
	VPField          VPType = 0x03
	VPValMethod      VPType = 0x04
	VPPtrMethod      VPType = 0x05
	VPInterface      VPType = 0x06
	VPSubrefField    VPType = 0x07 // not deref type
	VPDerefField     VPType = 0x13 // 0x10 + VPField
	VPDerefValMethod VPType = 0x14 // 0x10 + VPValMethod
	VPDerefPtrMethod VPType = 0x15 // 0x10 + VPPtrMethod
	VPDerefInterface VPType = 0x16 // 0x10 + VPInterface
	// 0x3X, 0x5X, 0x7X, 0x9X, 0xAX, 0xCX, 0xEX reserved.
)

func NewValuePath(t VPType, depth uint8, index uint16, n Name) ValuePath {
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
	return NewValuePath(VPUverse, 0, index, n)
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

func (vp ValuePath) Validate() {
	vp.validateDepth()

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

// ----------------------------------------
// Utility

func (x *BasicLitExpr) GetString() string {
	// Matches string literal parsing in go/constant.MakeFromLiteral.
	str, err := strconv.Unquote(x.Value)
	if err != nil {
		panic("error in parsing string literal: " + err.Error())
	}
	return str
}

func (x *BasicLitExpr) GetInt() int {
	i, err := strconv.Atoi(x.Value)
	if err != nil {
		panic(err)
	}
	return i
}

var rePkgName = regexp.MustCompile(`^[a-z][a-z0-9_]+$`)

// TODO: consider length restrictions.
// If this function is changed, ReadMemPackage's documentation should be updated accordingly.
func validatePkgName(name Name) error {
	if !rePkgName.MatchString(string(name)) {
		return fmt.Errorf("invalid package name %q", name)
	}
	return nil
}

// The distinction is used for validation to work
// both before and after preprocessing.
const (
	missingResultNamePrefix    = ".res." // if there was no name
	underscoreResultNamePrefix = ".res_" // if was underscore
)

//nolint:unused
func isUnnamedResult(name Name) bool {
	return isMissingResult(name) || isUnderscoreResult(name)
}

func isMissingResult(name Name) bool {
	return strings.HasPrefix(string(name), missingResultNamePrefix)
}

//nolint:unused
func isUnderscoreResult(name Name) bool {
	return strings.HasPrefix(string(name), underscoreResultNamePrefix)
}
