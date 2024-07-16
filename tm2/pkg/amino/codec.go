package amino

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/amino/pkg"
)

// Useful for debugging.
const printLog = false

// ----------------------------------------
// Codec internals

type TypeInfo struct {
	Type      reflect.Type // never a pointer kind.
	Package   *Package     // package associated with Type.
	PtrToType reflect.Type
	ZeroValue reflect.Value
	InterfaceInfo
	ConcreteInfo
	StructInfo
}

type InterfaceInfo struct{}

type ConcreteInfo struct {
	Registered            bool      // Registered with Register*().
	Name                  string    // Registered name which may override default reflection name.
	PointerPreferred      bool      // Deserialize to pointer type if possible.
	TypeURL               string    // <domain and path>/<p3 package no slashes>.<Name>
	IsAminoMarshaler      bool      // Implements MarshalAmino() (<ReprObject>, error) and UnmarshalAmino(<ReprObject>) (error).
	ReprType              *TypeInfo // <ReprType> if IsAminoMarshaler, that, or by default the identity Type.
	IsJSONValueType       bool      // If true, the Any representation uses the "value" field (instead of embedding @type).
	IsBinaryWellKnownType bool      // If true, use built-in functions to encode/decode.
	IsJSONWellKnownType   bool      // If true, use built-in functions to encode/decode.
	IsJSONAnyValueType    bool      // If true, the interface/Any representation uses the "value" field.
	Elem                  *TypeInfo // Set if Type.Kind() is Slice or Array.
	ElemIsPtr             bool      // Set true iff Type.Elem().Kind() is Pointer.
}

type StructInfo struct {
	Fields []FieldInfo // If a struct.
}

type FieldInfo struct {
	Type         reflect.Type  // Struct field reflect.Type.
	TypeInfo     *TypeInfo     // Dereferenced struct field TypeInfo
	Name         string        // Struct field name
	Index        int           // Struct field index
	ZeroValue    reflect.Value // Could be nil pointer unlike TypeInfo.ZeroValue.
	UnpackedList bool          // True iff this field should be encoded as an unpacked list.
	FieldOptions               // Encoding options
}

type FieldOptions struct {
	JSONName      string // (JSON) field name
	JSONOmitEmpty bool   // (JSON) omitempty
	BinFixed64    bool   // (Binary) Encode as fixed64
	BinFixed32    bool   // (Binary) Encode as fixed32
	BinFieldNum   uint32 // (Binary) max 1<<29-1

	Unsafe         bool // e.g. if this field is a float.
	WriteEmpty     bool // write empty structs and lists (default false except for pointers)
	NilElements    bool // Empty list elements are decoded as nil iff set, otherwise are never nil.
	UseGoogleTypes bool // If true, decodes Any timestamp and duration to google types.
}

// ----------------------------------------
// TypeInfo convenience

func (info *TypeInfo) GetTyp3(fopts FieldOptions) Typ3 {
	return typeToTyp3(info.ReprType.Type, fopts)
}

// Used to determine whether to create an implicit struct or not.  Notice that
// the binary encoding of a list to be unpacked is indistinguishable from a
// struct that contains that list.
// NOTE: we expect info.Elem to be prepopulated, constructed within the scope
// of a Codec.
func (info *TypeInfo) IsStructOrUnpacked(fopt FieldOptions) bool {
	rinfo := info.ReprType
	if rinfo.Type.Kind() == reflect.Struct || rinfo.Type.Kind() == reflect.Interface {
		return true
	}
	// We can't just look at the kind and info.Type.Elem(),
	// as for example, a []time.Duration should not be packed,
	// but should be represented as a slice of structs.
	// For these cases, we should expect info.Elem to be prepopulated.
	if rinfo.Type.Kind() == reflect.Array || rinfo.Type.Kind() == reflect.Slice {
		return rinfo.Elem.GetTyp3(fopt) == Typ3ByteLength
	}
	return false
}

// If this is a slice or array, get .Elem.ReprType until no longer slice or
// array.
func (info *TypeInfo) GetUltimateElem() *TypeInfo {
	if info.Elem != nil {
		return info.Elem.ReprType.GetUltimateElem()
	}
	return info
}

func (info *TypeInfo) String() string {
	if info.Type == nil {
		// since we set it on the codec map
		// before it's fully populated.
		return "<new TypeInfo>"
	}
	buf := new(bytes.Buffer)
	buf.Write([]byte("TypeInfo{"))
	buf.Write([]byte(fmt.Sprintf("Type:%v,", info.Type)))
	if info.ConcreteInfo.Registered {
		buf.Write([]byte("Registered:true,"))
		buf.Write([]byte(fmt.Sprintf("PointerPreferred:%v,", info.PointerPreferred)))
		buf.Write([]byte(fmt.Sprintf("TypeURL:\"%v\",", info.TypeURL)))
	} else {
		buf.Write([]byte("Registered:false,"))
	}
	if info.ReprType == info {
		buf.Write([]byte(fmt.Sprintf("ReprType:<self>,")))
	} else {
		buf.Write([]byte(fmt.Sprintf("ReprType:\"%v\",", info.ReprType)))
	}
	if info.Type.Kind() == reflect.Struct {
		buf.Write([]byte(fmt.Sprintf("Fields:%v,", info.Fields)))
	}
	buf.Write([]byte("}"))
	return buf.String()
}

// ----------------------------------------
// FieldInfo convenience

func (finfo *FieldInfo) IsPtr() bool {
	return finfo.Type.Kind() == reflect.Ptr
}

func (finfo *FieldInfo) ValidateBasic() {
	if finfo.BinFixed32 {
		switch finfo.TypeInfo.GetUltimateElem().Type.Kind() {
		case reflect.Int32, reflect.Uint32:
			// ok
		case reflect.Int, reflect.Uint:
			// TODO error upon overflow/underflow during conversion.
			panic("\"fixed32\" not yet supported for int/uint")
		default:
			panic("unexpected tag \"fixed32\" for non-32bit type")
		}
	}
	if finfo.BinFixed64 {
		switch finfo.TypeInfo.GetUltimateElem().Type.Kind() {
		case reflect.Int64, reflect.Uint64, reflect.Int, reflect.Uint:
			// ok
		default:
			panic("unexpected tag \"fixed64\" for non-64bit type")
		}
	}
	if !finfo.Unsafe {
		switch finfo.TypeInfo.Type.Kind() {
		case reflect.Float32, reflect.Float64:
			panic("floating point types are unsafe for go-amino")
		}
		switch finfo.TypeInfo.GetUltimateElem().Type.Kind() {
		case reflect.Float32, reflect.Float64:
			panic("floating point types are unsafe for go-amino, even for repr types")
		}
	}
}

// ----------------------------------------
// Codec

type Codec struct {
	mtx       sync.RWMutex
	sealed    bool
	autoseal  bool
	typeInfos map[reflect.Type]*TypeInfo
	// proto3 name of format "<pkg path no slashes>.<MessageName>"
	// which follows the TypeURL's last (and required) slash.
	// only registered types have names.
	fullnameToTypeInfo map[string]*TypeInfo
	packages           pkg.PackageSet
	usePBBindings      bool
}

func NewCodec() *Codec {
	cdc := &Codec{
		sealed:             false,
		autoseal:           false,
		typeInfos:          make(map[reflect.Type]*TypeInfo),
		fullnameToTypeInfo: make(map[string]*TypeInfo),
		packages:           pkg.NewPackageSet(),
		usePBBindings:      false,
	}
	cdc.registerWellKnownTypes()
	return cdc
}

// Returns a new codec that is optimized w/ pbbindings.
// The returned codec is sealed, but may be affected by
// modifications to the underlying codec.
func (cdc *Codec) WithPBBindings() *Codec {
	return &Codec{
		sealed:             cdc.sealed,
		autoseal:           cdc.autoseal,
		typeInfos:          cdc.typeInfos,
		fullnameToTypeInfo: cdc.fullnameToTypeInfo,
		packages:           cdc.packages,
		usePBBindings:      true,
	}
}

// The package isn't (yet) necessary besides to get the full name of concrete
// types.  Registers all dependencies of pkg recursively.  This operation is
// idempotent -- pkgs already registered may be registered again.
func (cdc *Codec) RegisterPackage(pkg *Package) {
	cdc.assertNotSealed()

	// Register dependencies if needed.
	for _, dep := range pkg.Dependencies {
		cdc.RegisterPackage(dep)
	}

	// Register types for package.
	for _, t := range pkg.Types {
		cdc.RegisterTypeFrom(t.Type, pkg)
	}
}

// This function should be used to register concrete types that will appear in
// interface fields/elements to be encoded/decoded by go-amino.
// You may want to use RegisterPackage() instead which registers everything in
// a package.
// Usage:
// `amino.RegisterTypeFrom(MyStruct1{}, "/tm.cryp.MyStruct1")`
func (cdc *Codec) RegisterTypeFrom(rt reflect.Type, pkg *Package) {
	cdc.assertNotSealed()

	// Get p3 full name.
	t, ok := pkg.GetType(rt)
	if !ok {
		panic(fmt.Errorf("type %v not registered on package %v", rt, pkg))
	}

	// Get type_url
	typeURL := pkg.TypeURLForType(rt)
	pointerPreferred := t.PointerPreferred
	cdc.registerType(pkg, rt, typeURL, pointerPreferred, true)
}

// This function exists so that typeURL etc can be overridden.
func (cdc *Codec) registerType(pkg *Package, rt reflect.Type, typeURL string, pointerPreferred bool, primary bool) {
	cdc.assertNotSealed()

	// Add package to packages if new.
	cdc.packages.Add(pkg)

	if rt.Kind() == reflect.Interface ||
		rt.Kind() == reflect.Ptr {
		panic(fmt.Sprintf("expected non-interface non-pointer concrete type, got %v", rt))
	}

	// Construct TypeInfo if one doesn't already exist.
	info, ok := cdc.typeInfos[rt]
	if ok {
		if info.Registered {
			// If idempotent operation, ignore silently.
			// Otherwise, panic.
			if info.Package != pkg {
				panic(fmt.Sprintf("type %v already registered with different package %v", rt, info.Package))
			}
			if info.ConcreteInfo.PointerPreferred != pointerPreferred {
				panic(fmt.Sprintf("type %v already registered with different pointer preference", rt))
			}
			if info.ConcreteInfo.TypeURL != typeURL {
				panic(fmt.Sprintf("type %v already registered with different type URL %v", rt, info.TypeURL))
			}
			return // silently
		} else {
			// we will be filling in an existing type.
		}
	} else {
		// construct a new one.
		info = cdc.newTypeInfoUnregisteredWLock(rt)
	}

	// Fill info for registered types.
	info.Package = pkg
	info.ConcreteInfo.Registered = true
	info.ConcreteInfo.PointerPreferred = pointerPreferred
	info.ConcreteInfo.Name = typeURLtoShortname(typeURL)
	info.ConcreteInfo.TypeURL = typeURL

	// Separate locking instance,
	// do the registration
	func() { // So it unlocks after scope.
		cdc.mtx.Lock()
		defer cdc.mtx.Unlock()
		cdc.registerTypeInfoWLocked(info, primary)
	}()

	func() { // And do it again...
		cdc.mtx.Lock()
		defer cdc.mtx.Unlock()
		// Cuz why not.
	}()
}

func (cdc *Codec) Seal() *Codec {
	cdc.mtx.Lock()
	defer cdc.mtx.Unlock()

	cdc.sealed = true
	return cdc
}

func (cdc *Codec) Autoseal() *Codec {
	cdc.mtx.Lock()
	defer cdc.mtx.Unlock()

	if cdc.sealed {
		panic("already sealed")
	}
	cdc.autoseal = true
	return cdc
}

// PrintTypes writes all registered types in a markdown-style table.
// The table's header is:
//
// | Type  | TypeURL | Notes |
//
// Where Type is the golang type name and TypeURL is the type_url the type was registered with.
func (cdc *Codec) PrintTypes(out io.Writer) error {
	cdc.mtx.RLock()
	defer cdc.mtx.RUnlock()
	// print header
	if _, err := io.WriteString(out, "| Type | TypeURL | Length | Notes |\n"); err != nil {
		return err
	}
	if _, err := io.WriteString(out, "| ---- | ------- | ------ | ----- |\n"); err != nil {
		return err
	}
	// only print concrete types for now (if we want everything, we can iterate over the typeInfos map instead)
	for _, i := range cdc.typeInfos {
		if _, err := io.WriteString(out, "| "); err != nil {
			return err
		}
		// TODO(ismail): optionally create a link to code on github:
		if _, err := io.WriteString(out, i.Type.Name()); err != nil {
			return err
		}
		if _, err := io.WriteString(out, " | "); err != nil {
			return err
		}
		if _, err := io.WriteString(out, i.TypeURL); err != nil {
			return err
		}
		if _, err := io.WriteString(out, " | "); err != nil {
			return err
		}

		if _, err := io.WriteString(out, getLengthStr(i)); err != nil {
			return err
		}

		if _, err := io.WriteString(out, " | "); err != nil {
			return err
		}
		// empty notes table data by default // TODO(ismail): make this configurable

		if _, err := io.WriteString(out, " |\n"); err != nil {
			return err
		}
	}
	// finish table
	return nil
}

// A heuristic to guess the size of a registered type and return it as a string.
// If the size is not fixed it returns "variable".
func getLengthStr(info *TypeInfo) string {
	switch info.Type.Kind() {
	case reflect.Array,
		reflect.Int8,
		reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:
		s := info.Type.Size()
		return fmt.Sprintf("0x%X", s)
	default:
		return "variable"
	}
}

// ----------------------------------------

func (cdc *Codec) assertNotSealed() {
	cdc.mtx.Lock()
	defer cdc.mtx.Unlock()

	if cdc.sealed {
		panic("codec sealed")
	}
}

func (cdc *Codec) doAutoseal() {
	cdc.mtx.Lock()
	defer cdc.mtx.Unlock()

	if cdc.autoseal {
		cdc.sealed = true
		cdc.autoseal = false
	}
}

// assumes write lock is held.
// primary should generally be true, and must be true for the first type set
// here that is info.Registered, except when registering secondary types for a
// given (full) name, such as google.protobuf.*.  If primary is set to false
// and info.Registered, the name must already be
// registered, and no side effects occur.
// CONTRACT: info.Type is set
// CONTRACT: if info.Registered, info.TypeURL is set
func (cdc *Codec) registerTypeInfoWLocked(info *TypeInfo, primary bool) {
	if info.Type.Kind() == reflect.Ptr {
		panic(fmt.Sprintf("unexpected pointer type"))
	}
	if existing, ok := cdc.typeInfos[info.Type]; !ok || existing != info {
		if !ok {
			// See corresponding comment in newTypeInfoUnregisteredWLocked.
			panic("unrecognized *TypeInfo")
		} else {
			panic(fmt.Sprintf("unexpected *TypeInfo: existing: %v, new: %v", existing, info))
		}
	}
	if !info.Registered {
		panic("expected registered info")
	}

	// Everybody's dooing a brand-new dance, now
	// Come on baby, doo the registration!
	fullname := typeURLtoFullname(info.TypeURL)
	existing, ok := cdc.fullnameToTypeInfo[fullname]
	if primary {
		if ok {
			panic(fmt.Sprintf("fullname <%s> already registered for %v (TypeURL: %v)", fullname, existing.Type, info.TypeURL))
		}
		cdc.fullnameToTypeInfo[fullname] = info
	} else {
		if !ok {
			panic(fmt.Sprintf("fullname <%s> not yet registered", fullname))
		}
	}
}

// XXX TODO: make this safe so modifications don't affect runtime codec,
// and ensure that it stays safe.
// NOTE: do not modify the returned Packages.
func (cdc *Codec) GetPackages() pkg.PackageSet {
	cdc.mtx.RLock()
	defer cdc.mtx.RUnlock()

	return cdc.packages
}

// This is used primarily for gengo.
// XXX TODO: make this safe so modifications don't affect runtime codec,
// and ensure that it stays safe.
// NOTE: do not modify the returned TypeInfo.
func (cdc *Codec) GetTypeInfo(rt reflect.Type) (info *TypeInfo, err error) {
	return cdc.getTypeInfoWLock(rt)
}

func (cdc *Codec) getTypeInfoWLock(rt reflect.Type) (info *TypeInfo, err error) {
	cdc.mtx.Lock() // requires wlock because we might set.
	// NOTE: We must defer, or at least recover, otherwise panics in
	// getTypeInfoWLocked() will render the codec locked.
	defer cdc.mtx.Unlock()

	info, err = cdc.getTypeInfoWLocked(rt)
	return info, err
}

// If a new one is constructed and cached in state, it is not yet registered.
// Automatically dereferences rt pointers.
func (cdc *Codec) getTypeInfoWLocked(rt reflect.Type) (info *TypeInfo, err error) {
	// Dereference pointer type.
	for rt.Kind() == reflect.Ptr {
		if rt.Elem().Kind() == reflect.Ptr {
			return nil, fmt.Errorf("cannot support nested pointers, got %v", rt)
		}
		rt = rt.Elem()
	}

	info, ok := cdc.typeInfos[rt]
	if !ok {
		info = cdc.newTypeInfoUnregisteredWLocked(rt)
	}
	return info, nil
}

func (cdc *Codec) getTypeInfoFromTypeURLRLock(typeURL string, fopts FieldOptions) (info *TypeInfo, err error) {
	fullname := typeURLtoFullname(typeURL)
	return cdc.getTypeInfoFromFullnameRLock(fullname, fopts)
}

func (cdc *Codec) getTypeInfoFromFullnameRLock(fullname string, fopts FieldOptions) (info *TypeInfo, err error) {
	// We do not use defer cdc.mtx.Unlock() here due to performance overhead of
	// defer in go1.11 (and prior versions). Ensure new code paths unlock the
	// mutex.
	cdc.mtx.RLock()

	// Special cases: time and duration
	if fullname == "google.protobuf.Timestamp" && !fopts.UseGoogleTypes {
		cdc.mtx.RUnlock()
		info, err = cdc.getTypeInfoWLock(timeType)
		return
	}
	if fullname == "google.protobuf.Duration" && !fopts.UseGoogleTypes {
		cdc.mtx.RUnlock()
		info, err = cdc.getTypeInfoWLock(durationType)
		return
	}

	info, ok := cdc.fullnameToTypeInfo[fullname]
	if !ok {
		err = fmt.Errorf("amino: unrecognized concrete type full name %s", fullname)
		cdc.mtx.RUnlock()
		return
	}
	cdc.mtx.RUnlock()
	return
}

// ----------------------------------------
// TypeInfo registration

// Constructs a *TypeInfo from scratch (except
// dependencies).  The constructed TypeInfo is stored in
// state, but not yet registered - no name or decoding
// preferece (pointer or not) is known, so it cannot be
// used to decode into an interface.
//
// cdc.registerType() calls this first for
// initial construction.  Unregistered type infos can
// still represent circular types because they still
// populate the internal lookup map, but they don't have
// certain fields set, such as:
//
//   - .Package - defaults to nil until registered.
//   - .ConcreteInfo.PointerPreferred - how it prefers to
//     be decoded
//   - .ConcreteInfo.TypeURL - for Any serialization
//
// But it does set .ConcreteInfo.Elem, which may be
// modified by the Codec instance.
func (cdc *Codec) newTypeInfoUnregisteredWLock(rt reflect.Type) *TypeInfo {
	cdc.mtx.Lock()
	defer cdc.mtx.Unlock()

	return cdc.newTypeInfoUnregisteredWLocked(rt)
}

func (cdc *Codec) newTypeInfoUnregisteredWLocked(rt reflect.Type) *TypeInfo {
	switch rt.Kind() {
	case reflect.Ptr:
		panic(fmt.Sprintf("unexpected pointer type %v", rt)) // should not happen.
	case reflect.Map:
		panic(fmt.Sprintf("map type not supported %v", rt))
	case reflect.Func:
		panic(fmt.Sprintf("func type not supported %v", rt))
	}
	if _, exists := cdc.typeInfos[rt]; exists {
		panic(fmt.Sprintf("type info already registered for %v", rt))
	}

	// Populate this early so it gets found when getTypeInfoWLocked() is
	// called, esp for parseStructInfoWLocked() which may cause infinite
	// recursion if two structs reference each other in declaration.
	// TODO: can protobuf support this? If not, we would still want to, but
	// restrict what can be compiled to protobuf, or something.
	info := new(TypeInfo)
	if _, exists := cdc.typeInfos[rt]; exists {
		panic("should not happen, instance already exists")
	}
	cdc.typeInfos[rt] = info

	info.Type = rt
	info.PtrToType = reflect.PointerTo(rt)
	info.ZeroValue = reflect.Zero(rt)
	var isAminoMarshaler bool
	var reprType reflect.Type
	if rm, ok := rt.MethodByName("MarshalAmino"); ok {
		isAminoMarshaler = true
		reprType = marshalAminoReprType(rm)
	}
	if rm, ok := reflect.PointerTo(rt).MethodByName("UnmarshalAmino"); ok {
		if !isAminoMarshaler {
			panic("Must implement both (o).MarshalAmino and (*o).UnmarshalAmino")
		}
		reprType2 := unmarshalAminoReprType(rm)
		if reprType != reprType2 {
			panic("Must match MarshalAmino and UnmarshalAmino repr types")
		}
	}
	/*
		NOTE: this could used by genproto typeToP3Type,
		but it isn't quite right... we don't want them to
		preserve the "Time" name, we want "Timestamp".

		// Special cases for well known types.
		// TODO: refactor out and merge into wellknown.go somehow.
		// NOTE: isAminoMarshaler remains false.
		switch rt {
		case timeType:
			reprType = gTimestampType
		case durationType:
			reprType = gDurationType
		}
		// END Special cases for well known types.
	*/
	if isAminoMarshaler {
		info.ConcreteInfo.IsAminoMarshaler = true
		rinfo, err := cdc.getTypeInfoWLocked(reprType)
		if err != nil {
			panic(err)
		}
		info.ConcreteInfo.ReprType = rinfo
	} else {
		info.ConcreteInfo.IsAminoMarshaler = false
		info.ConcreteInfo.ReprType = info
	}
	info.ConcreteInfo.IsBinaryWellKnownType = isBinaryWellKnownType(rt)
	info.ConcreteInfo.IsJSONWellKnownType = isJSONWellKnownType(rt)
	info.ConcreteInfo.IsJSONAnyValueType = isJSONAnyValueType(rt)
	if rt.Kind() == reflect.Array || rt.Kind() == reflect.Slice {
		einfo, err := cdc.getTypeInfoWLocked(rt.Elem())
		if err != nil {
			panic(err)
		}
		info.ConcreteInfo.Elem = einfo
		info.ConcreteInfo.ElemIsPtr = rt.Elem().Kind() == reflect.Ptr
	}
	if rt.Kind() == reflect.Struct {
		info.StructInfo = cdc.parseStructInfoWLocked(rt)
	}
	return info
}

// ----------------------------------------
// ...

func (cdc *Codec) parseStructInfoWLocked(rt reflect.Type) (sinfo StructInfo) {
	defer func() {
		if ex := recover(); ex != nil {
			panic(fmt.Sprintf("panic parsing struct %v",
				rt))
		}
	}()
	if rt.Kind() != reflect.Struct {
		panic("should not happen")
	}

	infos := make([]FieldInfo, 0, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		ftype := field.Type
		if !isExported(field) {
			continue // field is unexported
		}
		skip, fopts := parseFieldOptions(field)
		if skip {
			continue // e.g. json:"-"
		}
		// NOTE: This is going to change a bit.
		// NOTE: BinFieldNum starts with 1.
		fopts.BinFieldNum = uint32(len(infos) + 1)
		fieldTypeInfo, err := cdc.getTypeInfoWLocked(ftype)
		if err != nil {
			panic(err)
		}
		frepr := fieldTypeInfo.ReprType.Type
		unpackedList := false
		if frepr.Kind() == reflect.Array || frepr.Kind() == reflect.Slice {
			if frepr.Elem().Kind() == reflect.Uint8 {
				// These get handled by our optimized methods,
				// encodeReflectBinaryByte[Slice/Array].
				unpackedList = false
			} else {
				etype := frepr.Elem()
				for etype.Kind() == reflect.Ptr {
					etype = etype.Elem()
				}
				typ3 := typeToTyp3(etype, fopts)
				if typ3 == Typ3ByteLength {
					unpackedList = true
				}
			}
		}

		fieldInfo := FieldInfo{
			Type:         ftype,
			TypeInfo:     fieldTypeInfo,
			Name:         field.Name, // Mostly for debugging.
			Index:        i,          // the field number for this go runtime (for decoding).
			ZeroValue:    reflect.Zero(ftype),
			UnpackedList: unpackedList,
			FieldOptions: fopts,
		}
		fieldInfo.ValidateBasic()
		infos = append(infos, fieldInfo)
	}
	sinfo = StructInfo{infos}
	return sinfo
}

func parseFieldOptions(field reflect.StructField) (skip bool, fopts FieldOptions) {
	binTag := field.Tag.Get("binary")
	aminoTag := field.Tag.Get("amino")
	jsonTag := field.Tag.Get("json")

	// If `json:"-"`, don't encode.
	// NOTE: This skips binary as well.
	if jsonTag == "-" {
		skip = true
		return
	}

	// Get JSON field name.
	jsonTagParts := strings.Split(jsonTag, ",")
	if jsonTagParts[0] == "" {
		fopts.JSONName = field.Name
	} else {
		fopts.JSONName = jsonTagParts[0]
	}

	// Get JSON omitempty.
	if len(jsonTagParts) > 1 {
		if jsonTagParts[1] == "omitempty" {
			fopts.JSONOmitEmpty = true
		}
	}

	// Parse binary tags.
	// NOTE: these get validated later, we don't have TypeInfo yet.
	if binTag == "fixed64" {
		fopts.BinFixed64 = true
	} else if binTag == "fixed32" {
		fopts.BinFixed32 = true
	}

	// Parse amino tags.
	aminoTags := strings.Split(aminoTag, ",")
	for _, aminoTag := range aminoTags {
		if aminoTag == "unsafe" {
			fopts.Unsafe = true
		}
		if aminoTag == "write_empty" {
			fopts.WriteEmpty = true
		}
		if aminoTag == "nil_elements" {
			fopts.NilElements = true
		}
	}

	return skip, fopts
}

// ----------------------------------------
// Misc.

func typeURLtoFullname(typeURL string) (fullname string) {
	parts := strings.Split(typeURL, "/")
	if len(parts) == 1 {
		panic(fmt.Sprintf("invalid type_url \"%v\", must contain at least one slash and be followed by the full name", typeURL))
	}
	return parts[len(parts)-1]
}

func typeURLtoShortname(typeURL string) (name string) {
	fullname := typeURLtoFullname(typeURL)
	parts := strings.Split(fullname, ".")
	if len(parts) == 1 {
		panic(fmt.Sprintf("invalid type_url \"%v\", full name must contain dot", typeURL))
	}
	return parts[len(parts)-1]
}

func typeToTyp3(rt reflect.Type, opts FieldOptions) Typ3 {
	// Special non-list cases:
	switch rt {
	case timeType:
		return Typ3ByteLength // for completeness
	case durationType:
		return Typ3ByteLength // as a google.protobuf.Duration.
	}
	// General cases:
	switch rt.Kind() {
	case reflect.Interface:
		return Typ3ByteLength
	case reflect.Array, reflect.Slice:
		return Typ3ByteLength
	case reflect.String:
		return Typ3ByteLength
	case reflect.Struct, reflect.Map:
		return Typ3ByteLength
	case reflect.Int64, reflect.Uint64:
		if opts.BinFixed64 {
			return Typ38Byte
		}
		return Typ3Varint
	case reflect.Int32, reflect.Uint32:
		if opts.BinFixed32 {
			return Typ34Byte
		}
		return Typ3Varint

	case reflect.Int16, reflect.Int8, reflect.Int,
		reflect.Uint16, reflect.Uint8, reflect.Uint, reflect.Bool:
		return Typ3Varint
	case reflect.Float64:
		return Typ38Byte
	case reflect.Float32:
		return Typ34Byte
	default:
		panic(fmt.Sprintf("unsupported field type %v", rt))
	}
}
