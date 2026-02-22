package gnolang

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/gnovm/pkg/gnolang",
	"gno",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(

	//----------------------------------------
	// Values
	&TypedValue{},
	StringValue{},
	BigintValue{},
	BigdecValue{},
	// DataByteValue{}
	PointerValue{},
	&ArrayValue{},
	&SliceValue{},
	&StructValue{},
	&FuncValue{},
	&MapValue{},
	&MapList{},
	&MapListItem{},
	&BoundMethodValue{},
	TypeValue{},
	&PackageValue{},
	&Block{},
	RefValue{},
	&HeapItemValue{},

	//----------------------------------------
	// Realm/Object
	ObjectID{},
	&ObjectInfo{},

	//----------------------------------------
	// Hash/Image
	ValueHash{},
	Hashlet{},

	//----------------------------------------
	// Nodes
	ValuePath{},
	Location{},
	// Name(""),
	Attributes{},
	NameExpr{},
	BasicLitExpr{},
	BinaryExpr{},
	CallExpr{},
	IndexExpr{},
	SelectorExpr{},
	SliceExpr{},
	StarExpr{},
	RefExpr{},
	TypeAssertExpr{},
	UnaryExpr{},
	CompositeLitExpr{},
	KeyValueExpr{},
	FuncLitExpr{},
	ConstExpr{},
	FieldTypeExpr{},
	ArrayTypeExpr{},
	SliceTypeExpr{},
	InterfaceTypeExpr{},
	ChanTypeExpr{},
	FuncTypeExpr{},
	MapTypeExpr{},
	StructTypeExpr{},
	constTypeExpr{},
	AssignStmt{},
	BlockStmt{},
	BranchStmt{},
	DeclStmt{},
	DeferStmt{},
	ExprStmt{},
	ForStmt{},
	GoStmt{},
	IfStmt{},
	IfCaseStmt{},
	IncDecStmt{},
	RangeStmt{},
	ReturnStmt{},
	SelectStmt{},
	SelectCaseStmt{},
	SendStmt{},
	SwitchStmt{},
	SwitchClauseStmt{},
	EmptyStmt{},
	bodyStmt{},
	FuncDecl{},
	ImportDecl{},
	ValueDecl{},
	TypeDecl{},

	//----------------------------------------
	// Nodes cont...
	StaticBlock{},
	FileSet{},
	FileNode{},
	PackageNode{},
	RefNode{},
	NameSource{},
	Pos{},
	Span{},

	//----------------------------------------
	// Types
	PrimitiveType(0),
	&PointerType{},
	&ArrayType{},
	&SliceType{},
	&StructType{},
	FieldType{},
	&FuncType{},
	&MapType{},
	&InterfaceType{},
	&TypeType{},
	&DeclaredType{},
	&PackageType{},
	&ChanType{},
	blockType{},
	&tupleType{},
	RefType{},
	heapItemType{},

	//----------------------------------------
	// MemPackage related
	MemPackageType(""),
	MemPackageFilter(""),
))
