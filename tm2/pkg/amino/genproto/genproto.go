package genproto

// p3c.SetProjectRootGopkg("example.com/main")

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/genproto/stringutil"
	"github.com/gnolang/gno/tm2/pkg/amino/pkg"
)

// TODO sort
//  * Proto3 import file paths are by default always full (including
//  domain) and basically the gopkg path.  This lets proto3 schema
//  import paths stay consistent even as dependency.
//  * In the go mod world, the user is expected to run an independent
//  tool to copy proto files to a proto folder from go mod dependencies.
//  This is provided by MakeProtoFilder().

// P3Context holds contextual information beyond the P3Doc.
//
// It holds all the package infos needed to derive the full P3doc,
// including p3 import paths, as well as where to write them,
// because all of that information is encapsulated in amino.Package.
//
// It also holds a local amino.Codec instance, with package registrations
// passed through.
type P3Context struct {
	// e.g. "github.com/tendermint/tendermint/abci/types" ->
	//   &Package{...}
	packages pkg.PackageSet

	// TODO
	// // for beyond default "type.proto"
	// // e.g. "tendermint.abci.types" ->
	// //   []string{"github.com/tendermint/abci/types/types.proto"}}
	// moreP3Imports map[string][]string

	// This is only necessary to construct TypeInfo.
	cdc *amino.Codec
}

func NewP3Context() *P3Context {
	p3c := &P3Context{
		packages: pkg.NewPackageSet(),
		cdc:      amino.NewCodec(),
	}
	return p3c
}

func (p3c *P3Context) RegisterPackage(pkg *amino.Package) {
	pkgs := pkg.CrawlPackages(nil)
	for _, pkg := range pkgs {
		p3c.registerPackage(pkg)
	}
}

func (p3c *P3Context) registerPackage(pkg *amino.Package) {
	p3c.packages.Add(pkg)
	p3c.cdc.RegisterPackage(pkg)
}

func (p3c *P3Context) GetPackage(gopkg string) *amino.Package {
	return p3c.packages.Get(gopkg)
}

// Crawls the packages and flattens all dependencies.
// Includes
func (p3c *P3Context) GetAllPackages() (res []*amino.Package) {
	seen := map[*amino.Package]struct{}{}
	for _, pkg := range p3c.packages {
		pkgs := pkg.CrawlPackages(seen)
		res = append(res, pkgs...)
	}
	for _, pkg := range p3c.cdc.GetPackages() {
		if _, exists := seen[pkg]; !exists {
			res = append(res, pkg)
		}
	}
	return
}

func (p3c *P3Context) ValidateBasic() {
	// TODO: do verifications across packages.
	// pkgs := p3c.GetAllPackages()
}

// TODO: This could live as a method of the package, and only crawl the
// dependencies of that package.  But a method implemented on P3Context
// should function like this and print an intelligent error.
// Set implicit to false to assert-that name matches in package.
// Set implicit to true for implicit structures like nested lists.
func (p3c *P3Context) GetP3ImportPath(p3type P3Type, implicit bool) string {
	p3pkg := p3type.GetPackageName()
	pkgs := p3c.GetAllPackages()
	for _, pkg := range pkgs {
		if pkg.P3PkgName == p3pkg {
			if implicit {
				return pkg.P3ImportPath
			} else if pkg.HasName(p3type.GetName()) {
				return pkg.P3ImportPath
			}
		}
	}
	panic(fmt.Sprintf("proto3 type %v not recognized", p3type))
}

// Given a codec and some reflection type, generate the Proto3 message
// (partial) schema.  Imports are added to p3doc.
func (p3c *P3Context) GenerateProto3MessagePartial(p3doc *P3Doc, rt reflect.Type) (p3msg P3Message) {
	if p3doc.PackageName == "" {
		panic("cannot generate message partials in the root package \"\".")
	}
	if rt.Kind() == reflect.Ptr {
		panic("pointers not yet supported. if you meant pointer-preferred (for decoding), pass in rt.Elem()")
	}
	if rt.Kind() == reflect.Interface {
		panic("nothing to generate for interfaces")
	}

	info, err := p3c.cdc.GetTypeInfo(rt)
	if err != nil {
		panic(err)
	}

	// The p3 schema is determined by the structure of ReprType.  But the name,
	// package, and where the binding artifacts get written, are all of the
	// original package.  Thus, .ReprType.Type.Name() and
	// .ReprType.Type.Package etc should not be used, and sometimes we must
	// preserve the original info's package as arguments along with .ReprType.
	rinfo := info.ReprType
	if rinfo.ReprType != rinfo {
		// info.ReprType should point to itself, chaining is not allowed.
		panic("should not happen")
	}

	rsfields := []amino.FieldInfo(nil)
	if rinfo.Type.Kind() == reflect.Struct {
		switch rinfo.Type {
		case timeType:
			// special case: time
			rinfo, err := p3c.cdc.GetTypeInfo(gTimestampType)
			if err != nil {
				panic(err)
			}
			rsfields = rinfo.StructInfo.Fields
		case durationType:
			// special case: duration
			rinfo, err := p3c.cdc.GetTypeInfo(gDurationType)
			if err != nil {
				panic(err)
			}
			rsfields = rinfo.StructInfo.Fields
		default:
			// general case
			rsfields = rinfo.StructInfo.Fields
		}
	} else {
		// implicit struct.
		// TODO: shouldn't this name end with "Wrapper" suffix?
		rsfields = []amino.FieldInfo{{
			Type:     rinfo.Type,
			TypeInfo: rinfo,
			Name:     "Value",
			FieldOptions: amino.FieldOptions{
				// TODO can we override JSON to unwrap here?
				BinFieldNum: 1,
			},
		}}
	}

	// When fields include other declared structs,
	// we need to know whether it's an external reference
	// (with corresponding imports in the proto3 schema)
	// or an internal reference (with no imports necessary).
	pkgPath := rt.PkgPath()
	if pkgPath == "" {
		panic(fmt.Errorf("can only generate proto3 message schemas from user-defined package-level declared structs, got rt %v", rt))
	}

	p3msg.Name = info.Name // not rinfo.

	var fieldComments map[string]string
	if rinfo.Package != nil {
		if pkgType, ok := rinfo.Package.GetType(rt); ok {
			p3msg.Comment = pkgType.Comment
			// We will check for optional field comments below.
			fieldComments = pkgType.FieldComments
		}
	}

	// Append to p3msg.Fields, fields of the struct.
	for _, field := range rsfields { // rinfo.
		fp3, fp3IsRepeated, implicit := typeToP3Type(info.Package, field.TypeInfo, field.FieldOptions)
		// If the p3 field package is the same, omit the prefix.
		if fp3.GetPackageName() == p3doc.PackageName {
			fp3m := fp3.(P3MessageType)
			fp3m.SetOmitPackage()
			fp3 = fp3m
		} else if fp3.GetPackageName() != "" {
			importPath := p3c.GetP3ImportPath(fp3, implicit)
			p3doc.AddImport(importPath)
		}
		p3Field := P3Field{
			Repeated: fp3IsRepeated,
			Type:     fp3,
			Name:     stringutil.ToLowerSnakeCase(field.Name),
			JSONName: field.JSONName,
			Number:   field.FieldOptions.BinFieldNum,
		}
		if fieldComments != nil {
			p3Field.Comment = fieldComments[field.Name]
		}
		p3msg.Fields = append(p3msg.Fields, p3Field)
	}

	return
}

// Generate the Proto3 message (partial) schema for an implist list.  Imports
// are added to p3doc.
func (p3c *P3Context) GenerateProto3ListPartial(p3doc *P3Doc, nl NList) (p3msg P3Message) {
	if p3doc.PackageName == "" {
		panic("cannot generate message partials in the root package \"\".")
	}

	ep3 := nl.ElemP3Type()
	if ep3.GetPackageName() == p3doc.PackageName {
		ep3m := ep3.(P3MessageType)
		ep3m.SetOmitPackage()
		ep3 = ep3m
	}
	p3Field := P3Field{
		Repeated: true,
		Type:     ep3,
		Name:     "Value",
		Number:   1,
	}
	p3msg.Name = nl.Name()
	p3msg.Fields = append(p3msg.Fields, p3Field)
	return
}

// Given the arguments, create a new P3Doc.
// pkg is optional.
func (p3c *P3Context) GenerateProto3SchemaForTypes(pkg *amino.Package, rtz ...reflect.Type) (p3doc P3Doc) {
	if pkg.P3PkgName == "" {
		panic(errors.New("cannot generate schema in the root package \"\""))
	}

	// Set the package.
	p3doc.PackageName = pkg.P3PkgName
	p3doc.GoPackage = pkg.P3GoPkgPath

	// Add declared imports.
	for _, dep := range pkg.GetAllDependencies() {
		p3doc.AddImport(dep.P3ImportPath)
	}

	// Set Message schemas.
	for _, rt := range rtz {
		if rt.Kind() == reflect.Interface {
			continue
		} else if rt.Kind() == reflect.Ptr {
			rt = rt.Elem()
		}
		p3msg := p3c.GenerateProto3MessagePartial(&p3doc, rt)
		p3doc.Messages = append(p3doc.Messages, p3msg)
	}

	// Collect list types and uniq,
	// then create list message schemas.
	// These are representational
	nestedListTypes := make(map[string]NList)
	for _, rt := range rtz {
		if rt.Kind() == reflect.Interface {
			continue
		}
		info, err := p3c.cdc.GetTypeInfo(rt)
		if err != nil {
			panic(err)
		}
		findNLists(pkg, info, &nestedListTypes)
	}
	for _, nl := range sortFound(nestedListTypes) {
		p3msg := p3c.GenerateProto3ListPartial(&p3doc, nl)
		p3doc.Messages = append(p3doc.Messages, p3msg)
	}

	return p3doc
}

// Convenience.
func (p3c *P3Context) WriteProto3SchemaForTypes(filename string, pkg *amino.Package, rtz ...reflect.Type) {
	fmt.Printf("writing proto3 schema to %v for package %v\n", filename, pkg)
	p3doc := p3c.GenerateProto3SchemaForTypes(pkg, rtz...)
	err := os.WriteFile(filename, []byte(p3doc.Print()), 0o644)
	if err != nil {
		panic(err)
	}
}

var (
	timeType     = reflect.TypeOf(time.Now())
	durationType = reflect.TypeOf(time.Duration(0))
)

// If info.ReprType is a struct, the returned proto3 type is a P3MessageType.
func typeToP3Type(root *amino.Package, info *amino.TypeInfo, fopts amino.FieldOptions) (p3type P3Type, repeated bool, implicit bool) {
	// Special case overrides.
	// We don't handle the case when info.ReprType.Type is time here.
	switch info.Type {
	case timeType:
		return NewP3MessageType("google.protobuf", "Timestamp"), false, false
	case durationType:
		return NewP3MessageType("google.protobuf", "Duration"), false, false
	}

	// Dereference type, in case pointer.
	rt := info.ReprType.Type
	switch rt.Kind() {
	case reflect.Interface:
		return P3AnyType, false, false
	case reflect.Bool:
		return P3ScalarTypeBool, false, false
	case reflect.Int:
		if fopts.BinFixed64 {
			return P3ScalarTypeSfixed64, false, false
		} else if fopts.BinFixed32 {
			return P3ScalarTypeSfixed32, false, false
		} else {
			return P3ScalarTypeSint64, false, false
		}
	case reflect.Int8:
		return P3ScalarTypeSint32, false, false
	case reflect.Int16:
		return P3ScalarTypeSint32, false, false
	case reflect.Int32:
		if fopts.BinFixed32 {
			return P3ScalarTypeSfixed32, false, false
		} else {
			return P3ScalarTypeSint32, false, false
		}
	case reflect.Int64:
		if fopts.BinFixed64 {
			return P3ScalarTypeSfixed64, false, false
		} else {
			return P3ScalarTypeSint64, false, false
		}
	case reflect.Uint:
		if fopts.BinFixed64 {
			return P3ScalarTypeFixed64, false, false
		} else if fopts.BinFixed32 {
			return P3ScalarTypeFixed32, false, false
		} else {
			return P3ScalarTypeUint64, false, false
		}
	case reflect.Uint8:
		return P3ScalarTypeUint32, false, false
	case reflect.Uint16:
		return P3ScalarTypeUint32, false, false
	case reflect.Uint32:
		if fopts.BinFixed32 {
			return P3ScalarTypeFixed32, false, false
		} else {
			return P3ScalarTypeUint32, false, false
		}
	case reflect.Uint64:
		if fopts.BinFixed64 {
			return P3ScalarTypeFixed64, false, false
		} else {
			return P3ScalarTypeUint64, false, false
		}
	case reflect.Float32:
		return P3ScalarTypeFloat, false, false
	case reflect.Float64:
		return P3ScalarTypeDouble, false, false
	case reflect.Complex64, reflect.Complex128:
		panic("complex types not yet supported")
	case reflect.Array, reflect.Slice:
		switch info.Elem.ReprType.Type.Kind() {
		case reflect.Uint8:
			return P3ScalarTypeBytes, false, false
		default:
			elemP3Type, elemRepeated, _ := typeToP3Type(root, info.Elem, fopts)
			if elemRepeated {
				elemP3Type = newNList(root, info, fopts).ElemP3Type()
				return elemP3Type, true, true
			}
			return elemP3Type, true, false
		}
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr,
		reflect.UnsafePointer:
		panic("chan, func, map, and pointers are not supported")
	case reflect.String:
		return P3ScalarTypeString, false, false
	case reflect.Struct:
		if info.Package == nil {
			panic(fmt.Sprintf("type %v not registered with codec", info.Type.Name()))
		}
		// NOTE: we don't use rt, because the p3 package and name should still
		// match the declaration, rather than inherit or refer to the repr type
		// (if it is registered at all).
		return NewP3MessageType(info.Package.P3PkgName, info.Name), false, false
	default:
		panic("unexpected rt kind")
	}
}

// Writes in the same directory as the origin package.
func WriteProto3Schema(pkg *amino.Package) {
	p3c := NewP3Context()
	p3c.RegisterPackage(pkg)
	p3c.ValidateBasic()
	filename := path.Join(pkg.DirName, pkg.GoPkgName+".proto")
	p3c.WriteProto3SchemaForTypes(filename, pkg, pkg.ReflectTypes()...)
}

// Symlinks .proto files from pkg info to dirname, keeping the go path
// structure as expected, <dirName>/path/to/gopkg/<gopkgName>.proto.
// If Pkg.DirName is empty, the package is considered "well known", and
// the mapping is not made.
func MakeProtoFolder(pkg *amino.Package, dirName string) {
	fmt.Printf("making proto3 schema folder for package %v\n", pkg)
	p3c := NewP3Context()
	p3c.RegisterPackage(pkg)

	// Populate mapping.
	// p3 import path -> p3 import file (abs path).
	// e.g. "github.com/.../mygopkg.proto" ->
	// "/gopath/pkg/mod/.../mygopkg.proto"
	p3imports := map[string]string{}
	for _, dpkg := range p3c.GetAllPackages() {
		if dpkg.P3SchemaFile == "" {
			// Skip well known packages like google.protobuf.Any
			continue
		}
		p3path := dpkg.P3ImportPath
		if p3path == "" {
			panic("P3ImportPath cannot be empty")
		}
		p3file := dpkg.P3SchemaFile
		p3imports[p3path] = p3file
	}

	// Check validity.
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		panic(fmt.Sprintf("directory %v does not exist", dirName))
	}

	// Make symlinks.
	for p3path, p3file := range p3imports {
		fmt.Println("p3path", p3path, "p3file", p3file)
		loc := path.Join(dirName, p3path)
		locdir := path.Dir(loc)
		// Ensure that paths exist.
		if _, err := os.Stat(locdir); os.IsNotExist(err) {
			err = os.MkdirAll(locdir, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}
		// Delete existing symlink.
		if _, err := os.Stat(loc); !os.IsNotExist(err) {
			err := os.Remove(loc)
			if err != nil {
				panic(err)
			}
		}
		// Write symlink.
		err := os.Symlink(p3file, loc)
		if os.IsExist(err) {
			// do nothing.
		} else if err != nil {
			panic(err)
		}
	}
}

// Uses pkg.P3GoPkgPath to determine where the compiled file goes.  If
// pkg.P3GoPkgPath is a subpath of pkg.GoPkgPath, then it will be
// written in the relevant subpath in pkg.DirName.
// `protosDir`: folder where .proto files for all dependencies live.
func RunProtoc(pkg *amino.Package, protosDir string) {
	if !strings.HasSuffix(pkg.P3SchemaFile, ".proto") {
		panic(fmt.Sprintf("expected P3Importfile to have .proto suffix, got %v", pkg.P3SchemaFile))
	}
	inDir := filepath.Dir(pkg.P3SchemaFile)
	inFile := filepath.Base(pkg.P3SchemaFile)
	outDir := path.Join(inDir, "pb")
	outFile := inFile[:len(inFile)-6] + ".pb.go"
	// Ensure that paths exist.
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		err = os.MkdirAll(outDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	// First generate output to a temp dir.
	tempDir, err := os.MkdirTemp("", "amino-genproto")
	if err != nil {
		return
	}
	// Run protoc
	cmd := exec.Command("protoc", "-I="+inDir, "-I="+protosDir, "--go_out="+tempDir, pkg.P3SchemaFile)
	fmt.Println("running protoc: ", cmd.String())
	cmd.Stdin = nil
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	if err != nil {
		fmt.Println("ERROR: ", out.String())
		panic(err)
	}

	// Copy file from tempDir to outDir.
	copyFile(
		path.Join(tempDir, pkg.P3GoPkgPath, outFile),
		path.Join(outDir, outFile),
	)
}

func copyFile(src string, dst string) {
	// Read all content of src to data
	data, err := os.ReadFile(src)
	if err != nil {
		panic(err)
	}
	// Write data to dst
	err = os.WriteFile(dst, data, 0o644)
	if err != nil {
		panic(err)
	}
}
