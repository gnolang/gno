package genproto

import (
	"fmt"
	"go/ast"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/libs/press"
)

//----------------------------------------

// NOTE: The goal is not complete Proto3 compatibility (unless there is
// widespread demand for maintaining this repo for that purpose).  Rather, the
// point is to define enough such that the subset that is needed for Amino
// Go->Proto3 is supported.  For example, there is explicitly no plan to
// support the automatic conversion of Proto3->Go, so not all features need to
// be supported.
// NOTE: enums are not supported, as Amino's philosophy is that value checking
// should primarily be done on the application side.

type P3Type interface {
	AssertIsP3Type()
	GetPackageName() string // proto3 package prefix
	GetName() string        // proto3 name
	GetFullName() string    // proto3 full name
}

func (P3ScalarType) AssertIsP3Type()  {}
func (P3MessageType) AssertIsP3Type() {}

type P3ScalarType string

func (P3ScalarType) GetPackageName() string { return "" }
func (st P3ScalarType) GetName() string     { return string(st) }
func (st P3ScalarType) GetFullName() string { return string(st) }

const (
	P3ScalarTypeDouble   P3ScalarType = "double"
	P3ScalarTypeFloat    P3ScalarType = "float"
	P3ScalarTypeInt32    P3ScalarType = "int32"
	P3ScalarTypeInt64    P3ScalarType = "int64"
	P3ScalarTypeUint32   P3ScalarType = "uint32"
	P3ScalarTypeUint64   P3ScalarType = "uint64"
	P3ScalarTypeSint32   P3ScalarType = "sint32"
	P3ScalarTypeSint64   P3ScalarType = "sint64"
	P3ScalarTypeFixed32  P3ScalarType = "fixed32"
	P3ScalarTypeFixed64  P3ScalarType = "fixed64"
	P3ScalarTypeSfixed32 P3ScalarType = "sfixed32"
	P3ScalarTypeSfixed64 P3ScalarType = "sfixed64"
	P3ScalarTypeBool     P3ScalarType = "bool"
	P3ScalarTypeString   P3ScalarType = "string"
	P3ScalarTypeBytes    P3ScalarType = "bytes"
)

type P3MessageType struct {
	PackageName string // proto3 package name, optional.
	Name        string // message name.
	OmitPackage bool   // if true, PackageName is not printed.
}

func NewP3MessageType(pkg string, name string) P3MessageType {
	if name == string(P3ScalarTypeDouble) ||
		name == string(P3ScalarTypeFloat) ||
		name == string(P3ScalarTypeInt32) ||
		name == string(P3ScalarTypeInt64) ||
		name == string(P3ScalarTypeUint32) ||
		name == string(P3ScalarTypeUint64) ||
		name == string(P3ScalarTypeSint32) ||
		name == string(P3ScalarTypeSint64) ||
		name == string(P3ScalarTypeFixed32) ||
		name == string(P3ScalarTypeFixed64) ||
		name == string(P3ScalarTypeSfixed32) ||
		name == string(P3ScalarTypeSfixed64) ||
		name == string(P3ScalarTypeBool) ||
		name == string(P3ScalarTypeString) ||
		name == string(P3ScalarTypeBytes) {
		panic(fmt.Sprintf("field type %v already defined", name))
	}
	// check name
	if len(name) == 0 {
		panic("custom p3 type name can't be empty")
	}
	return P3MessageType{PackageName: pkg, Name: name}
}

var P3AnyType P3MessageType = NewP3MessageType("google.protobuf", "Any")

// May be empty if it isn't set (for locally declared messages).
func (p3mt P3MessageType) GetPackageName() string {
	return p3mt.PackageName
}

func (p3mt P3MessageType) GetName() string {
	return p3mt.Name
}

func (p3mt P3MessageType) GetFullName() string {
	if p3mt.OmitPackage || p3mt.PackageName == "" {
		return p3mt.Name
	} else {
		return fmt.Sprintf("%v.%v", p3mt.PackageName, p3mt.Name)
	}
}

func (p3mt *P3MessageType) SetOmitPackage() {
	p3mt.OmitPackage = true
}

func (p3mt P3MessageType) String() string {
	return p3mt.GetFullName()
}

// NOTE: P3Doc and its fields are meant to hold basic AST-like information.  No
// validity checking happens here... it should happen before these values are
// set.  Convenience functions that require much more context like P3Context are OK.
type P3Doc struct {
	PackageName string
	GoPackage   string // TODO replace with general options
	Comment     string
	Imports     []P3Import
	Messages    []P3Message
	// Enums []P3Enums // enums not supported, no need.
}

func (doc *P3Doc) AddImport(path string) {
	for _, p3import := range doc.Imports {
		if p3import.Path == path {
			return // do nothing.
		}
	}
	doc.Imports = append(doc.Imports, P3Import{Path: path})
}

type P3Import struct {
	Path string
	// Public bool // not used (yet)
}

type P3Message struct {
	Comment string
	Name    string
	Fields  []P3Field
}

type P3Field struct {
	Comment  string
	Repeated bool
	Type     P3Type
	Name     string
	JSONName string
	Number   uint32
}

//----------------------------------------
// Functions for printing P3 objects

// NOTE: P3Doc imports must be set correctly.
func (doc P3Doc) Print() string {
	p := press.NewPress()
	return strings.TrimSpace(doc.PrintCode(p).Print())
}

func (doc P3Doc) PrintCode(p *press.Press) *press.Press {
	p.Pl("syntax = \"proto3\";")
	if doc.PackageName != "" {
		p.Pl("package %v;", doc.PackageName)
	}
	// Print comments, if any.
	p.Ln()
	if doc.Comment != "" {
		printComments(p, doc.Comment)
		p.Ln()
	}
	// Print options, if any.
	if doc.GoPackage != "" {
		p.Pl("option go_package = \"%v\";", doc.GoPackage)
		p.Ln()
	}
	// Print imports, if any.
	for i, imp := range doc.Imports {
		if i == 0 {
			p.Pl("// imports")
		}
		imp.PrintCode(p)
		if i == len(doc.Imports)-1 {
			p.Ln()
		}
	}
	// Print message schemas, if any.
	for i, msg := range doc.Messages {
		if i == 0 {
			p.Pl("// messages")
		}
		msg.PrintCode(p)
		p.Ln()
		if i == len(doc.Messages)-1 {
			p.Ln()
		}
	}
	return p
}

func (imp P3Import) PrintCode(p *press.Press) *press.Press {
	p.Pl("import %v;", strconv.Quote(imp.Path))
	return p
}

func (msg P3Message) Print() string {
	p := press.NewPress()
	return msg.PrintCode(p).Print()
}

func (msg P3Message) PrintCode(p *press.Press) *press.Press {
	printComments(p, msg.Comment)
	p.Pl("message %v {", msg.Name).I(func(p *press.Press) {
		for _, fld := range msg.Fields {
			fld.PrintCode(p)
		}
	}).Pl("}")
	return p
}

func (fld P3Field) PrintCode(p *press.Press) *press.Press {
	fieldOptions := ""
	if fld.JSONName != "" && fld.JSONName != fld.Name {
		fieldOptions = " [json_name = \"" + fld.JSONName + "\"]"
	}
	printComments(p, fld.Comment)
	if fld.Repeated {
		p.Pl("repeated %v %v = %v%v;", fld.Type, fld.Name, fld.Number, fieldOptions)
	} else {
		p.Pl("%v %v = %v%v;", fld.Type, fld.Name, fld.Number, fieldOptions)
	}
	return p
}

func printComments(p *press.Press, comment string) {
	if comment == "" {
		return
	}
	commentLines := strings.Split(comment, "\n")
	for _, line := range commentLines {
		p.Pl("// %v", line)
	}
}

//----------------------------------------
// Synthetic type for nested lists

// This exists as a workaround due to Proto deficiencies,
// namely how fields can only be repeated, not nestedly-repeated.
type NList struct {
	// Define dimension as followes:
	// []struct{} has dimension 1, as well as [][]byte.
	// [][]struct{} has dimension 2, as well as [][][]byte.
	// When dimension is 2 or greater, we need implicit structs.
	// The NestedType is meant to represent these types,
	// so Dimensions is usually 2 or greater.
	Dimensions int

	// UltiElem.ReprType might not be UltiElem.
	// Could be []byte.
	UltiElem *amino.TypeInfo

	// Optional Package, where this nested list was used.
	// NOTE: two packages can't (yet?) share nested lists.
	Package *amino.Package

	// If embedded in a struct.
	// Should be sanitized to uniq properly.
	FieldOptions amino.FieldOptions
}

// filter to field options that matter for NLists.
func nListFieldOptions(fopts amino.FieldOptions) amino.FieldOptions {
	return amino.FieldOptions{
		BinFixed64:     fopts.BinFixed64,
		BinFixed32:     fopts.BinFixed32,
		UseGoogleTypes: fopts.UseGoogleTypes,
	}
}

// info: a list's TypeInfo.
func newNList(pkg *amino.Package, info *amino.TypeInfo, fopts amino.FieldOptions) NList {
	if !isListType(info.ReprType.Type) {
		panic("should not happen")
	}
	if !isListType(info.ReprType.Type) {
		panic("should not happen")
	}
	if info.ReprType.Elem.ReprType.Type.Kind() == reflect.Uint8 {
		panic("should not happen")
	}
	fopts = nListFieldOptions(fopts)
	einfo := info
	leinfo := (*amino.TypeInfo)(nil)
	counter := 0
	for isListType(einfo.ReprType.Type) {
		leinfo = einfo
		einfo = einfo.ReprType.Elem
		counter++
	}
	if einfo.ReprType.Type.Name() == "uint8" {
		einfo = leinfo
		counter--
	}
	return NList{
		Package:      pkg,
		Dimensions:   counter,
		UltiElem:     einfo,
		FieldOptions: fopts,
	}
}

func (nl NList) Name() string {
	if nl.Dimensions <= 0 {
		panic("should not happen")
	}
	pkgname := strings.ToUpper(nl.Package.GoPkgName) // must be exposed.
	var prefix string
	var ename string
	listSfx := strings.Repeat("List", nl.Dimensions)

	ert := nl.UltiElem.ReprType.Type
	if isListType(ert) {
		if nl.UltiElem.ReprType.Elem.ReprType.Type.Kind() != reflect.Uint8 {
			panic("should not happen")
		}
		ename = "Bytes"
	} else {
		// Get name from .Type, not ReprType.Type.
		ename = nl.UltiElem.Name
	}

	if nl.FieldOptions.BinFixed64 {
		prefix = "Fixed64"
	} else if nl.FieldOptions.BinFixed32 {
		prefix = "Fixed32"
	}
	if nl.FieldOptions.UseGoogleTypes {
		prefix = "G" + prefix
	}

	return fmt.Sprintf("%s_%v%v%v", pkgname, prefix, ename, listSfx)
}

//nolint:staticcheck
func (nl NList) P3GoExprString(imports *ast.GenDecl, scope *ast.Scope) string {
	pkgName := addImportAuto(imports, scope, nl.Package.GoPkgName+"pb", nl.Package.P3GoPkgPath)
	return fmt.Sprintf("*%v.%v", pkgName, nl.Name())
}

// NOTE: requires nl.Package.
func (nl NList) P3Type() P3Type {
	return NewP3MessageType(
		nl.Package.P3PkgName,
		nl.Name(),
	)
}

func (nl NList) Elem() NList {
	if nl.Dimensions == 1 {
		panic("should not happen")
	}
	return NList{
		Package:      nl.Package,
		Dimensions:   nl.Dimensions - 1,
		UltiElem:     nl.UltiElem,
		FieldOptions: nl.FieldOptions,
	}
}

func (nl NList) ElemP3Type() P3Type {
	if nl.Dimensions == 1 {
		p3type, repeated, implicit := typeToP3Type(
			nl.Package,
			nl.UltiElem,
			nl.FieldOptions,
		)
		if repeated || implicit {
			panic("should not happen")
		}
		return p3type
	} else {
		return nl.Elem().P3Type()
	}
}

// For uniq'ing.
func (nl NList) Key() string {
	return fmt.Sprintf("%v.%v", nl.Package.GoPkgName, nl.Name())
}

//----------------------------------------
// Other

// Find root struct fields that are nested list types.
// If not a struct, assume an implicit struct with single field.
// If type is amino.Marshaler, find values/fields from the repr.
// Pointers are ignored, even for the terminal type.
// e.g. if TypeInfo.ReprType.Type is
//   - struct{ [][]int, [][]string } -> return [][]int, [][]string
//   - [][]int -> return [][]int
//   - [][][]int -> return [][][]int, [][]int
//   - [][][]byte -> return [][][]byte (but not [][]byte, which is just repeated bytes).
//   - [][][][]int -> return [][][][]int, [][][]int, [][]int.
//
// The results are uniq'd and sorted somehow.
func findNLists(root *amino.Package, info *amino.TypeInfo, found *map[string]NList) {
	if found == nil {
		*found = map[string]NList{}
	}
	switch info.ReprType.Type.Kind() {
	case reflect.Struct:
		for _, field := range info.ReprType.Fields {
			fert := field.TypeInfo.ReprType.Type
			fopts := field.FieldOptions
			if isListType(fert) {
				lists := findNLists2(root, field.TypeInfo, fopts)
				for _, list := range lists {
					if list.Dimensions >= 1 {
						(*found)[list.Key()] = list
					}
				}
			}
		}
		return
	case reflect.Array, reflect.Slice:
		lists := findNLists2(root, info, amino.FieldOptions{})
		for _, list := range lists {
			if list.Dimensions >= 2 {
				(*found)[list.Key()] = list
			}
		}
	}
}

// The last item of res is the deepest.
// As a special recursive case, may return Dimensions:1 for bytes.
func findNLists2(root *amino.Package, list *amino.TypeInfo, fopts amino.FieldOptions) []NList {
	fopts = nListFieldOptions(fopts)
	switch list.ReprType.Type.Kind() {
	case reflect.Ptr:
		panic("should not happen")
	case reflect.Array, reflect.Slice:
		elem := list.ReprType.Elem
		if isListType(elem.ReprType.Type) {
			if elem.ReprType.Elem.ReprType.Type.Kind() == reflect.Uint8 {
				// elem is []byte or bytes, and list is []bytes.
				// no need to look for sublists.
				return []NList{
					{
						Package:      root,
						Dimensions:   1,
						UltiElem:     elem,
						FieldOptions: fopts,
					},
				}
			} else {
				sublists := findNLists2(root, elem, fopts)
				if len(sublists) == 0 {
					return []NList{{
						Package:      root,
						Dimensions:   1,
						UltiElem:     elem.ReprType.Elem,
						FieldOptions: fopts,
					}}
				} else {
					deepest := sublists[len(sublists)-1]
					this := NList{
						Package:      root,
						Dimensions:   deepest.Dimensions + 1,
						UltiElem:     deepest.UltiElem,
						FieldOptions: fopts,
					}
					lists := append(sublists, this)
					return lists
				}
			}
		} else {
			return nil // nothing.
		}
	default:
		panic("should not happen")
	}
}

func sortFound(found map[string]NList) (res []NList) {
	for _, nl := range found {
		res = append(res, nl)
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i].Name() < res[j].Name() {
			return true
		} else if res[i].Name() == res[j].Name() {
			return res[i].Dimensions < res[j].Dimensions
		} else {
			return false
		}
	})
	return res
}
