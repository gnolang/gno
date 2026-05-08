package components

import (
	"encoding/hex"
	"fmt"
	"html/template"
	"math"
	"strconv"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
)

// StateNode is a UI-friendly decoded representation of a gno value, suitable
// for rendering as HTML. The walker (DecodePkgJSON, DecodeObjectJSON) consumes
// the raw Amino JSON returned by vm/qpkg_json or vm/qobject_json and produces
// a tree of StateNodes.
//
// Mirrors the StateNode contract from misc/gnojs (decode.ts), but built on
// native gnolang Go types via amino.UnmarshalJSON — no mirror types, no
// duplicated decoding logic.
type StateNode struct {
	// Name is the display label (variable name, struct field name, map key, slice index).
	Name string

	// Type is the human-readable type display (e.g. "int", "map[string]User").
	Type string

	// Kind is a simplified category used for styling/branching.
	Kind string

	// Value is the displayed leaf value (e.g. "42", "\"hello\"", "<cycle :1>").
	Value string

	// Expandable means the node has children that can be shown.
	Expandable bool

	// Children are inline children already decoded.
	Children []StateNode

	// ObjectID is set when this node is a stored object reference (RefValue).
	ObjectID string

	// TypeID is set when type information for this node may be useful.
	TypeID string

	// Length is the count of elements for collections.
	Length *int

	// Preview is a short one-line summary built from the first few
	// children — e.g. `{name: "alice", age: 30, …}` for a struct or
	// `{"general": Board, "dev": Board, …}` for a map. Renders next to
	// the type in collapsed/ref rows so users see what's behind the
	// arrow without expanding. Set by the walker for nodes with
	// decoded children, and re-computed by EnrichInlinePreviews after
	// stored refs are fetched lazily.
	Preview string

	// Source is the source-code location for functions and closures.
	Source *SourceLocation

	// SourceHTML, when set, is the chroma-highlighted source snippet for the
	// referenced span. The walker leaves this empty (it doesn't read files);
	// the orchestrator that wires the renderer fills it in. Typed as
	// template.HTML so html/template treats it as already-safe markup.
	SourceHTML template.HTML

	// Href, when non-empty, is the navigation URL pointing to this node's
	// own state-explorer page (only set when ObjectID is set, by the
	// orchestrator). Built via weburl.GnoURL so encoding stays consistent
	// with the rest of gnoweb. Typed template.URL so html/template trusts it.
	Href template.URL

	// OwnerHref is the navigation URL for OwnerID's own state page —
	// pre-built by the orchestrator so the audit-chip "Owner" link
	// preserves the page's height (time-travel) without the template
	// having to know about the URL syntax. Empty when OwnerID is.
	OwnerHref template.URL

	// Anchor, when non-empty, is the HTML id stamped on this node's
	// top-level row so the sidebar TOC can link to it via #fragment.
	// Set by Build{Package,Object}Sidebar — never by the walker.
	Anchor string

	// ---- ObjectInfo metadata (already in qobject_json/qpkg_json) ----
	// Captured by the walker for any node carrying an ObjectInfo: useful
	// for blockchain audit (Owner, Hash) and storage analysis (RefCount,
	// LastObjectSize) — without an extra fetch.

	// Hash is the content hash of the stored object (hex). Empty when
	// not applicable.
	Hash string

	// OwnerID is the ObjectID of the parent that owns this one in the
	// gnolang ownership tree — i.e. who can mutate it.
	OwnerID string

	// ModTime is the chain height at which this object was last modified.
	ModTime string

	// RefCount is the persistence-ref count for this object.
	RefCount string

	// LastObjectSize is the storage size in bytes (decimal string).
	LastObjectSize string

	// Doc is the Go-style documentation comment attached to the
	// declaration in source. Populated post-walk by the handler from
	// the package's JSON doc index, matched by Name. Markdown-formatted.
	Doc string
}

// StateObjectInfoView mirrors a stored object's ObjectInfo, formatted for
// display in the sidebar. The orchestrator extracts this from the queried
// object's outermost Value (qobject_json response).
type StateObjectInfoView struct {
	Hash, OwnerID, ModTime, RefCount, LastObjectSize string
	IsEscaped                                        bool
}

// DecodedObject is what handlers receive when calling DecodeObjectFull —
// the children to render plus the metadata to surface in the sidebar.
type DecodedObject struct {
	Nodes []StateNode
	Info  StateObjectInfoView
}

// SourceLocation pinpoints a span in a source file.
type SourceLocation struct {
	File      string
	StartLine int
	EndLine   int
}

// ---- Public entry points ----

type pkgResponse struct {
	Names  []string         `json:"names"`
	Values []gno.TypedValue `json:"values"`
}

type objectResponse struct {
	ObjectID string    `json:"objectid"`
	Value    gno.Value `json:"value"`
}

// DecodePkgJSON decodes a vm/qpkg_json response into top-level StateNodes.
func DecodePkgJSON(raw []byte) ([]StateNode, error) {
	var resp pkgResponse
	if err := amino.UnmarshalJSON(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode pkg JSON: %w", err)
	}
	nodes := make([]StateNode, 0, len(resp.Names))
	for i, name := range resp.Names {
		if i >= len(resp.Values) {
			break
		}
		nodes = append(nodes, decodeTypedValue(name, resp.Values[i]))
	}
	return nodes, nil
}

// DecodeObjectJSON decodes a vm/qobject_json response into the children
// StateNodes of the contained object. Without a type context, struct fields
// fall back to positional indices because Amino strips the named-type
// definition during ExportValues (DeclaredType → RefType, no Fields).
// Use DecodeObjectJSONWithType to recover field names.
func DecodeObjectJSON(raw []byte) ([]StateNode, error) {
	var resp objectResponse
	if err := amino.UnmarshalJSON(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode object JSON: %w", err)
	}
	return decodeValueChildren(resp.Value), nil
}

// DecodeObjectJSONWithType decodes a vm/qobject_json response together with
// a vm/qtype_json response so struct field names appear instead of "0", "1",
// "2" placeholders. The type response is what `qtype_json(<TypeID>)` returns
// for the object's named type — pass nil/empty rawType to fall back to plain
// DecodeObjectJSON behaviour.
//
// Field-name resolution applies to the queried object's TOP-LEVEL struct
// fields. Nested structs whose Type is itself a RefType keep positional
// indices; users see them resolved when they navigate into their own
// per-object pages (each carrying its own &tid=).
func DecodeObjectJSONWithType(rawObject, rawType []byte) ([]StateNode, error) {
	var resp objectResponse
	if err := amino.UnmarshalJSON(rawObject, &resp); err != nil {
		return nil, fmt.Errorf("decode object JSON: %w", err)
	}

	// No type context provided: fall back to indices.
	if len(rawType) == 0 {
		return decodeValueChildren(resp.Value), nil
	}

	var typeResp struct {
		TypeID string   `json:"typeid"`
		Type   gno.Type `json:"type"`
	}
	if err := amino.UnmarshalJSON(rawType, &typeResp); err != nil || typeResp.Type == nil {
		// Best-effort: bad type response shouldn't fail the page render.
		return decodeValueChildren(resp.Value), nil
	}

	return decodeValueChildrenTyped(resp.Value, typeResp.Type, typeResp.TypeID), nil
}

// decodeValueChildrenTyped is like decodeValueChildren but uses an outer
// gno.Type to resolve struct field names. The type comes from a separate
// qtype_json round-trip — Amino strips named-type field info from the
// value tree at export. The originalTid is propagated so that nested ref
// nodes (heap → ref pattern) carry it forward to round 2 of inline
// preview, which would otherwise fall back to positional indices.
func decodeValueChildrenTyped(v gno.Value, t gno.Type, originalTid string) []StateNode {
	// HeapItemValue: the type describes the inner TypedValue. Synthesize
	// {T: t, V: hiv.Value.V, N: hiv.Value.N} so decodeStruct sees the
	// resolved StructType when it walks fields.
	if hiv, ok := v.(*gno.HeapItemValue); ok {
		node := decodeTypedValue("value", gno.TypedValue{
			T: t, V: hiv.Value.V, N: hiv.Value.N,
		})
		// If the synthesis produced a ref node without a TypeID (typical
		// when t is an anonymous StructType), forward the original tid so
		// the next preview round still resolves field names.
		if node.ObjectID != "" && node.TypeID == "" && originalTid != "" {
			node.TypeID = originalTid
		}
		if len(node.Children) > 0 {
			return node.Children
		}
		return []StateNode{node}
	}

	// Direct StructValue under a known StructType: label fields directly.
	if sv, ok := v.(*gno.StructValue); ok {
		bt := baseType(t)
		if st, ok := bt.(*gno.StructType); ok {
			children := make([]StateNode, len(sv.Fields))
			for i, ftv := range sv.Fields {
				name := strconv.Itoa(i)
				if i < len(st.Fields) && st.Fields[i].Name != "" {
					name = string(st.Fields[i].Name)
				}
				children[i] = decodeTypedValue(name, ftv)
			}
			return children
		}
	}

	// Other shapes: use the existing un-typed children logic.
	return decodeValueChildren(v)
}

// ---- Core walker ----

// maxDecodeDepth caps recursion in the value walker. Adversarial or
// pathological values (deeply nested closures, cycles the cycle-marker
// missed, etc.) cannot drag the renderer into a stack overflow — at the
// cap the walker yields a sentinel "(too deep)" leaf. Generous enough
// that real Gno values never hit it.
const maxDecodeDepth = 256

// tooDeepNode is the sentinel emitted when a subtree exceeds maxDecodeDepth.
func tooDeepNode(name string) StateNode {
	return StateNode{Name: name, Type: "(too deep)", Kind: "truncated", Value: "…"}
}

func decodeTypedValue(name string, tv gno.TypedValue) StateNode {
	return decodeTypedValueAt(0, name, tv)
}

func decodeTypedValueAt(depth int, name string, tv gno.TypedValue) StateNode {
	if depth >= maxDecodeDepth {
		return tooDeepNode(name)
	}
	if tv.T == nil {
		return StateNode{Name: name, Type: "<nil>", Kind: "nil", Value: "nil"}
	}

	tName := typeName(tv.T)
	kind := typeKind(tv.T)
	bt := baseType(tv.T)
	typeID := getTypeID(tv.T)

	// FuncType stored as RefValue → expandable to fetch source on detail page.
	if _, isFunc := bt.(*gno.FuncType); isFunc {
		if rv, ok := tv.V.(gno.RefValue); ok {
			return StateNode{
				Name: name, Type: funcSignature(tv.T), Kind: "func",
				Expandable: true, ObjectID: rv.ObjectID.String(),
			}
		}
	}

	// RefValue: persisted object reference.
	if rv, ok := tv.V.(gno.RefValue); ok {
		if rv.PkgPath != "" {
			return StateNode{Name: name, Type: tName, Kind: "package", Value: rv.PkgPath}
		}
		return StateNode{
			Name: name, Type: tName, Kind: kind,
			Expandable: true, ObjectID: rv.ObjectID.String(), TypeID: typeID,
		}
	}

	// ExportRefValue: cycle-breaking marker. Kind is preserved from the type
	// (pointer, ref, …) — the value carries the cycle marker for styling.
	if erv, ok := tv.V.(gno.ExportRefValue); ok {
		return StateNode{
			Name: name, Type: tName, Kind: kind,
			Value: fmt.Sprintf("<cycle %s>", erv.ObjectID),
		}
	}

	// HeapItemValue: transparent unwrap (does not consume a depth slot —
	// it's a pass-through wrapper, not a structural level).
	if hiv, ok := tv.V.(*gno.HeapItemValue); ok {
		return decodeTypedValueAt(depth, name, hiv.Value)
	}

	// TypeValue: type definition shown as a value.
	if tv2, ok := tv.V.(gno.TypeValue); ok {
		return StateNode{Name: name, Type: "type", Kind: "type", Value: typeName(tv2.Type)}
	}

	// Primitives.
	if pt, ok := bt.(gno.PrimitiveType); ok {
		return decodePrimitive(name, tName, pt, tv)
	}

	// Struct (inline).
	if sv, ok := tv.V.(*gno.StructValue); ok {
		return decodeStruct(depth, name, tName, typeID, bt, sv)
	}

	// Array (inline).
	if av, ok := tv.V.(*gno.ArrayValue); ok {
		return decodeArray(depth, name, tName, av)
	}

	// Slice (inline or ref-based).
	if sv, ok := tv.V.(*gno.SliceValue); ok {
		return decodeSlice(depth, name, tName, sv)
	}

	// Map (inline).
	if mv, ok := tv.V.(*gno.MapValue); ok {
		return decodeMap(depth, name, tName, mv)
	}

	// Pointer (inline or to ref).
	if pv, ok := tv.V.(gno.PointerValue); ok {
		return decodePointer(depth, name, tName, typeID, pv)
	}

	// Func / Closure inline.
	if fv, ok := tv.V.(*gno.FuncValue); ok {
		return decodeFuncInline(depth, name, fv)
	}

	// Zero value (type but no value).
	if tv.V == nil {
		return StateNode{Name: name, Type: tName, Kind: kind, Value: "<zero>"}
	}

	// Fallback.
	return StateNode{Name: name, Type: tName, Kind: kind, Value: fmt.Sprintf("<%T>", tv.V)}
}

// decodeValueChildren turns a queried Object's Value (from vm/qobject_json)
// into the children to display on its dedicated page. For collection-shaped
// values we return one StateNode per field/element. For scalar-shaped values
// (a stored function, a pointer, a type value) we return a single StateNode
// representing the object itself — otherwise the page would render empty
// for any non-collection object.
func decodeValueChildren(v gno.Value) []StateNode {
	switch cv := v.(type) {
	case nil:
		return nil

	// ---- Collection-shaped: render fields/elements as direct children ----

	case *gno.StructValue:
		nodes := make([]StateNode, len(cv.Fields))
		for i, ftv := range cv.Fields {
			nodes[i] = decodeTypedValue(strconv.Itoa(i), ftv)
		}
		return nodes
	case *gno.ArrayValue:
		if cv.Data != nil {
			n := len(cv.Data)
			return []StateNode{{
				Name: "data", Type: "[]byte", Kind: "primitive",
				Value: fmt.Sprintf("[%d]byte{...}", n),
			}}
		}
		nodes := make([]StateNode, len(cv.List))
		for i, etv := range cv.List {
			nodes[i] = decodeTypedValue(strconv.Itoa(i), etv)
		}
		return nodes
	case *gno.SliceValue:
		// Slices live in their backing array; if the array is inline, expose
		// the visible window as elements. If the array is itself a stored
		// ref, expose a single ref node so the user can navigate.
		if av, ok := cv.Base.(*gno.ArrayValue); ok && av.Data == nil {
			offset, length := cv.Offset, cv.Length
			if offset < 0 {
				offset = 0
			}
			end := offset + length
			if end > len(av.List) {
				end = len(av.List)
			}
			nodes := make([]StateNode, end-offset)
			for i, etv := range av.List[offset:end] {
				nodes[i] = decodeTypedValue(strconv.Itoa(i), etv)
			}
			return nodes
		}
		if rv, ok := cv.Base.(gno.RefValue); ok {
			n := cv.Length
			return []StateNode{{
				Name: "(slice)", Type: "[]…", Kind: "slice",
				Expandable: n > 0, ObjectID: rv.ObjectID.String(), Length: intPtr(n),
			}}
		}
		return nil
	case *gno.MapValue:
		var nodes []StateNode
		if cv.List != nil {
			for cur := cv.List.Head; cur != nil; cur = cur.Next {
				keyStr := previewTypedValue(cur.Key)
				nodes = append(nodes, decodeTypedValue(keyStr, cur.Value))
			}
		}
		return nodes
	case *gno.HeapItemValue:
		// HeapItemValue is an internal wrapper for heap-allocated values
		// (the typical case for objects). When the inner value is itself
		// a collection (struct/array/map/slice), surface its children
		// directly so the page doesn't show a redundant "value :" row.
		// For scalars/funcs the inner becomes a single labelled row.
		inner := decodeTypedValue("value", cv.Value)
		if len(inner.Children) > 0 {
			return inner.Children
		}
		return []StateNode{inner}
	case *gno.Block:
		nodes := make([]StateNode, len(cv.Values))
		for i, tv := range cv.Values {
			nodes[i] = decodeTypedValue(strconv.Itoa(i), tv)
		}
		return nodes

	// ---- Scalar-shaped: wrap into a single representative node ----

	case *gno.FuncValue:
		// Top-level FuncValue (e.g. a stored package function): show the
		// function as one expandable node so the source bloc renders.
		name := string(cv.Name)
		if name == "" {
			name = "(function)"
		}
		return []StateNode{decodeFuncInline(0, name, cv)}
	case *gno.BoundMethodValue:
		return []StateNode{decodeFuncInline(0, "(method)", cv.Func)}
	case gno.PointerValue:
		// If it points at an inline TypedValue, surface that. If it points
		// at a stored ref, expose the navigation handle.
		if cv.TV != nil {
			return []StateNode{decodeTypedValue("*", *cv.TV)}
		}
		if rv, ok := cv.Base.(gno.RefValue); ok {
			return []StateNode{{
				Name: "*", Type: "(stored)", Kind: "pointer",
				Expandable: true, ObjectID: rv.ObjectID.String(),
			}}
		}
		return []StateNode{{Name: "*", Type: "*", Kind: "pointer", Value: "nil"}}
	case gno.TypeValue:
		return []StateNode{{
			Name: "(type)", Type: "type", Kind: "type",
			Value: typeName(cv.Type),
		}}
	case gno.RefValue:
		// The queried object resolved to a ref (rare at top level, but
		// possible). Surface it as a navigable handle so the user can drill
		// further rather than seeing an empty page.
		return []StateNode{{
			Name: "(ref)", Type: "(stored)", Kind: "ref",
			Expandable: true, ObjectID: cv.ObjectID.String(),
		}}
	case gno.ExportRefValue:
		return []StateNode{{
			Name: "(cycle)", Type: "(cycle)", Kind: "cycle",
			Value: fmt.Sprintf("<cycle %s>", cv.ObjectID),
		}}
	case gno.StringValue:
		return []StateNode{{
			Name: "(string)", Type: "string", Kind: "primitive",
			Value: quoteString(string(cv)),
		}}
	}
	return nil
}

// ---- Per-kind decoders ----

func decodePrimitive(name, tName string, pt gno.PrimitiveType, tv gno.TypedValue) StateNode {
	// Strings live in V (StringValue), all other primitives in N.
	if pt == gno.StringType || pt == gno.UntypedStringType {
		s := tv.GetString()
		return StateNode{Name: name, Type: tName, Kind: "primitive", Value: quoteString(s)}
	}

	// If V is missing and N is zero, render the zero value for the primitive.
	if tv.V == nil && tv.N == [8]byte{} {
		return StateNode{Name: name, Type: tName, Kind: "primitive", Value: zeroValueFor(pt)}
	}

	return StateNode{Name: name, Type: tName, Kind: "primitive", Value: primitiveDisplay(pt, tv)}
}

func decodeStruct(depth int, name, tName, typeID string, bt gno.Type, sv *gno.StructValue) StateNode {
	fieldNames := structFieldNames(bt)
	children := make([]StateNode, len(sv.Fields))
	for i, ftv := range sv.Fields {
		var fname string
		if i < len(fieldNames) && fieldNames[i] != "" {
			fname = fieldNames[i]
		} else {
			fname = strconv.Itoa(i)
		}
		children[i] = decodeTypedValueAt(depth+1, fname, ftv)
	}
	length := len(sv.Fields)
	node := StateNode{
		Name: name, Type: tName, Kind: "struct",
		Expandable: length > 0, Children: children,
		TypeID:  typeID,
		Length:  intPtr(length),
		Preview: buildChildrenPreview(children),
	}
	applyObjectInfo(&node, sv.ObjectInfo)
	return node
}

func decodeArray(depth int, name, tName string, av *gno.ArrayValue) StateNode {
	if av.Data != nil {
		n := len(av.Data)
		node := StateNode{
			Name: name, Type: tName, Kind: "array",
			Value:  fmt.Sprintf("[%d]byte{...}", n),
			Length: intPtr(n),
		}
		applyObjectInfo(&node, av.ObjectInfo)
		return node
	}
	children := make([]StateNode, len(av.List))
	for i, etv := range av.List {
		children[i] = decodeTypedValueAt(depth+1, strconv.Itoa(i), etv)
	}
	node := StateNode{
		Name: name, Type: tName, Kind: "array",
		Expandable: len(children) > 0, Children: children,
		Length: intPtr(len(av.List)),
	}
	applyObjectInfo(&node, av.ObjectInfo)
	return node
}

func decodeSlice(depth int, name, tName string, sv *gno.SliceValue) StateNode {
	length := sv.Length
	// Base is a RefValue → slice points at a stored array.
	if rv, ok := sv.Base.(gno.RefValue); ok {
		return StateNode{
			Name: name, Type: tName, Kind: "slice",
			Expandable: length > 0, ObjectID: rv.ObjectID.String(), Length: intPtr(length),
		}
	}
	// Base is an inline ArrayValue.
	if av, ok := sv.Base.(*gno.ArrayValue); ok {
		if av.Data != nil {
			return StateNode{
				Name: name, Type: tName, Kind: "slice",
				Value:  fmt.Sprintf("[]byte (len=%d)", length),
				Length: intPtr(length),
			}
		}
		offset := sv.Offset
		end := offset + length
		if end > len(av.List) {
			end = len(av.List)
		}
		if offset < 0 {
			offset = 0
		}
		visible := av.List[offset:end]
		children := make([]StateNode, len(visible))
		for i, etv := range visible {
			children[i] = decodeTypedValueAt(depth+1, strconv.Itoa(i), etv)
		}
		return StateNode{
			Name: name, Type: tName, Kind: "slice",
			Expandable: len(children) > 0, Children: children,
			Length: intPtr(length),
		}
	}
	return StateNode{
		Name: name, Type: tName, Kind: "slice",
		Expandable: length > 0, Length: intPtr(length),
	}
}

func decodeMap(depth int, name, tName string, mv *gno.MapValue) StateNode {
	var children []StateNode
	if mv.List != nil {
		for cur := mv.List.Head; cur != nil; cur = cur.Next {
			keyStr := previewTypedValue(cur.Key)
			children = append(children, decodeTypedValueAt(depth+1, keyStr, cur.Value))
		}
	}
	node := StateNode{
		Name: name, Type: tName, Kind: "map",
		Expandable: len(children) > 0, Children: children,
		Length:  intPtr(len(children)),
		Preview: buildChildrenPreview(children),
	}
	applyObjectInfo(&node, mv.ObjectInfo)
	return node
}

func decodePointer(depth int, name, tName, typeID string, pv gno.PointerValue) StateNode {
	if rv, ok := pv.Base.(gno.RefValue); ok {
		return StateNode{
			Name: name, Type: tName, Kind: "pointer",
			Expandable: true, ObjectID: rv.ObjectID.String(),
			// Carry the pointee's TypeID so inline preview / dedicated page
			// can fetch qtype_json and resolve struct field names.
			TypeID: typeID,
		}
	}
	if pv.TV != nil {
		child := decodeTypedValueAt(depth+1, "*", *pv.TV)
		return StateNode{
			Name: name, Type: tName, Kind: "pointer",
			Expandable: true, Children: []StateNode{child},
		}
	}
	return StateNode{Name: name, Type: tName, Kind: "pointer", Value: "nil"}
}

func decodeFuncInline(depth int, name string, fv *gno.FuncValue) StateNode {
	sig := "func()"
	if fv.Type != nil {
		sig = funcSignature(fv.Type)
	} else if fv.Name != "" {
		sig = fmt.Sprintf("func %s()", fv.Name)
	}
	src := extractFuncSource(fv)
	hasCaps := len(fv.Captures) > 0
	kind := "func"
	if hasCaps {
		kind = "closure"
	}
	if hasCaps {
		children := make([]StateNode, len(fv.Captures))
		for i, capture := range fv.Captures {
			children[i] = decodeTypedValueAt(depth+1, "value", capture)
		}
		return StateNode{
			Name: name, Type: sig, Kind: kind,
			Expandable: true, Source: src, Children: children,
		}
	}
	// Regular funcs: expandable when we have a Source range so the raw
	// tree view can disclose the body inline (matches the initial Jae
	// PR). Without Source they stay flat — nothing to disclose.
	return StateNode{
		Name: name, Type: sig, Kind: kind,
		Expandable: src != nil, Source: src,
	}
}

// ---- Type helpers (mirror type-utils.ts) ----

func typeName(t gno.Type) string {
	if t == nil {
		return "<nil>"
	}
	switch tt := t.(type) {
	case gno.PrimitiveType:
		return primitiveTypeName(tt)
	case *gno.PointerType:
		return "*" + typeName(tt.Elt)
	case *gno.ArrayType:
		return fmt.Sprintf("[%d]%s", tt.Len, typeName(tt.Elt))
	case *gno.SliceType:
		return "[]" + typeName(tt.Elt)
	case *gno.MapType:
		return fmt.Sprintf("map[%s]%s", typeName(tt.Key), typeName(tt.Value))
	case *gno.StructType:
		return "struct{...}"
	case *gno.FuncType:
		return "func(...)"
	case *gno.InterfaceType:
		return "interface{...}"
	case gno.RefType:
		id := tt.ID.String()
		dot := strings.LastIndex(id, ".")
		if dot >= 0 {
			pkgPath := id[:dot]
			parts := strings.Split(pkgPath, "/")
			return parts[len(parts)-1] + id[dot:]
		}
		return id
	case *gno.DeclaredType:
		parts := strings.Split(tt.PkgPath, "/")
		return parts[len(parts)-1] + "." + string(tt.Name)
	case *gno.TypeType:
		return "type"
	case *gno.PackageType:
		return "package"
	case *gno.ChanType:
		return "chan " + typeName(tt.Elt)
	}
	return fmt.Sprintf("<%T>", t)
}

func typeKind(t gno.Type) string {
	if t == nil {
		return "nil"
	}
	switch tt := t.(type) {
	case gno.PrimitiveType:
		return "primitive"
	case *gno.PointerType:
		return "pointer"
	case *gno.ArrayType:
		return "array"
	case *gno.SliceType:
		return "slice"
	case *gno.StructType:
		return "struct"
	case *gno.MapType:
		return "map"
	case *gno.FuncType:
		return "func"
	case *gno.InterfaceType:
		return "interface"
	case gno.RefType:
		return "ref"
	case *gno.DeclaredType:
		return typeKind(tt.Base)
	case *gno.TypeType:
		return "type"
	case *gno.PackageType:
		return "package"
	case *gno.ChanType:
		return "chan"
	}
	return "unknown"
}

// baseType unwraps DeclaredType to its underlying base type.
func baseType(t gno.Type) gno.Type {
	if dt, ok := t.(*gno.DeclaredType); ok {
		return dt.Base
	}
	return t
}

// structFieldNames returns the field names for a StructType (or DeclaredType
// wrapping one). Returns nil if the type isn't a struct.
func structFieldNames(t gno.Type) []string {
	switch tt := t.(type) {
	case *gno.StructType:
		names := make([]string, len(tt.Fields))
		for i, f := range tt.Fields {
			names[i] = string(f.Name)
		}
		return names
	case *gno.DeclaredType:
		return structFieldNames(tt.Base)
	}
	return nil
}

// getTypeID returns a TypeID suitable for `qtype_json` lookup. Drills
// through PointerType so `*foo.Bar` resolves to the inner Bar's TypeID —
// critical for inline preview to show field names on stored structs
// reached via a pointer (the typical realm pattern).
func getTypeID(t gno.Type) string {
	if t == nil {
		return ""
	}
	switch tt := t.(type) {
	case gno.RefType:
		return tt.ID.String()
	case *gno.DeclaredType:
		return tt.PkgPath + "." + string(tt.Name)
	case *gno.PointerType:
		return getTypeID(tt.Elt)
	}
	return ""
}

// funcSignature builds a human-readable signature, hiding the implicit `cur realm`
// crossing parameter when present (matches TS heuristic).
func funcSignature(t gno.Type) string {
	ft, ok := t.(*gno.FuncType)
	if !ok {
		return "func()"
	}
	params := make([]string, 0, len(ft.Params))
	for _, p := range ft.Params {
		// Hide implicit crossing param: name starts with "cur" and type is RefType.
		name := string(p.Name)
		if strings.HasPrefix(name, "cur") {
			if _, isRef := p.Type.(gno.RefType); isRef {
				continue
			}
		}
		tn := typeName(p.Type)
		if name != "" && !strings.HasPrefix(name, ".") {
			params = append(params, name+" "+tn)
		} else {
			params = append(params, tn)
		}
	}
	results := make([]string, 0, len(ft.Results))
	for _, r := range ft.Results {
		tn := typeName(r.Type)
		name := string(r.Name)
		if name != "" && !strings.HasPrefix(name, ".") {
			results = append(results, name+" "+tn)
		} else {
			results = append(results, tn)
		}
	}
	var ret string
	switch len(results) {
	case 0:
		ret = ""
	case 1:
		ret = " " + results[0]
	default:
		ret = " (" + strings.Join(results, ", ") + ")"
	}
	return "func(" + strings.Join(params, ", ") + ")" + ret
}

func extractFuncSource(fv *gno.FuncValue) *SourceLocation {
	if fv == nil || fv.Source == nil {
		return nil
	}
	rn, ok := fv.Source.(gno.RefNode)
	if !ok {
		return nil
	}
	if rn.Location.File == "" {
		return nil
	}
	return &SourceLocation{
		File:      rn.Location.File,
		StartLine: rn.Location.Span.Pos.Line,
		EndLine:   rn.Location.Span.End.Line,
	}
}

// ---- Primitive display ----

func primitiveTypeName(pt gno.PrimitiveType) string {
	switch pt {
	case gno.BoolType:
		return "bool"
	case gno.UntypedBoolType:
		return "untyped bool"
	case gno.StringType:
		return "string"
	case gno.UntypedStringType:
		return "untyped string"
	case gno.IntType:
		return "int"
	case gno.Int8Type:
		return "int8"
	case gno.Int16Type:
		return "int16"
	case gno.Int32Type:
		return "int32"
	case gno.Int64Type:
		return "int64"
	case gno.UintType:
		return "uint"
	case gno.Uint8Type:
		return "uint8"
	case gno.Uint16Type:
		return "uint16"
	case gno.Uint32Type:
		return "uint32"
	case gno.Uint64Type:
		return "uint64"
	case gno.Float32Type:
		return "float32"
	case gno.Float64Type:
		return "float64"
	case gno.UntypedRuneType:
		return "rune"
	case gno.UntypedBigintType:
		return "untyped bigint"
	case gno.UntypedBigdecType:
		return "untyped bigdec"
	case gno.DataByteType:
		return "databyte"
	}
	return fmt.Sprintf("prim(%d)", pt)
}

func primitiveDisplay(pt gno.PrimitiveType, tv gno.TypedValue) string {
	switch pt {
	case gno.BoolType, gno.UntypedBoolType:
		if tv.GetBool() {
			return "true"
		}
		return "false"
	case gno.IntType:
		return strconv.FormatInt(tv.GetInt(), 10)
	case gno.Int8Type:
		return strconv.FormatInt(int64(tv.GetInt8()), 10)
	case gno.Int16Type:
		return strconv.FormatInt(int64(tv.GetInt16()), 10)
	case gno.Int32Type, gno.UntypedRuneType:
		return strconv.FormatInt(int64(tv.GetInt32()), 10)
	case gno.Int64Type:
		return strconv.FormatInt(tv.GetInt64(), 10)
	case gno.UintType, gno.Uint64Type:
		return strconv.FormatUint(tv.GetUint64(), 10)
	case gno.Uint8Type, gno.DataByteType:
		return strconv.FormatUint(uint64(tv.GetUint8()), 10)
	case gno.Uint16Type:
		return strconv.FormatUint(uint64(tv.GetUint16()), 10)
	case gno.Uint32Type:
		return strconv.FormatUint(uint64(tv.GetUint32()), 10)
	case gno.Float32Type:
		return strconv.FormatFloat(float64(math.Float32frombits(tv.GetFloat32())), 'g', -1, 32)
	case gno.Float64Type:
		return strconv.FormatFloat(math.Float64frombits(tv.GetFloat64()), 'g', -1, 64)
	}
	return tv.String()
}

func zeroValueFor(pt gno.PrimitiveType) string {
	switch pt {
	case gno.BoolType, gno.UntypedBoolType:
		return "false"
	case gno.StringType, gno.UntypedStringType:
		return `""`
	}
	return "0"
}

// ---- String formatting ----

const stringTruncateLimit = 256

func quoteString(s string) string {
	if len(s) > stringTruncateLimit {
		return strconv.Quote(s[:stringTruncateLimit]) + "..."
	}
	return strconv.Quote(s)
}

// inlinePreviewMaxFields caps the number of children surfaced in the
// one-line preview of a collapsed struct or map. Three keeps the row
// readable; further fields collapse to "…".
const inlinePreviewMaxFields = 3

// buildChildrenPreview turns decoded children into a short one-liner
// like `{name: "alice", age: 30, …}` for rendering next to the type
// when the row is collapsed. Works uniformly for struct fields and map
// entries (the walker pre-formats each key as the child Name). Returns
// "" when there's nothing to show — the template branches on emptiness.
func buildChildrenPreview(children []StateNode) string {
	if len(children) == 0 {
		return ""
	}
	parts := make([]string, 0, inlinePreviewMaxFields)
	limit := len(children)
	if limit > inlinePreviewMaxFields {
		limit = inlinePreviewMaxFields
	}
	for i := 0; i < limit; i++ {
		parts = append(parts, children[i].Name+": "+previewChildValue(children[i]))
	}
	if len(children) > inlinePreviewMaxFields {
		parts = append(parts, "…")
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// previewChildValue picks the most compact representation of a child
// node for inline previews: the leaf value when present, the type
// (with length when known) otherwise. Truncates long strings.
func previewChildValue(c StateNode) string {
	if c.Value != "" {
		v := c.Value
		if len(v) > previewLimit {
			v = v[:previewLimit] + "…"
		}
		return v
	}
	if c.Length != nil {
		return fmt.Sprintf("%s(%d)", c.Type, *c.Length)
	}
	return c.Type
}

const previewLimit = 64

func previewTypedValue(tv gno.TypedValue) string {
	if tv.T == nil {
		return "nil"
	}
	bt := baseType(tv.T)
	if pt, ok := bt.(gno.PrimitiveType); ok {
		if pt == gno.StringType || pt == gno.UntypedStringType {
			s := tv.GetString()
			if len(s) > previewLimit {
				return strconv.Quote(s[:previewLimit]) + "..."
			}
			return strconv.Quote(s)
		}
		return primitiveDisplay(pt, tv)
	}
	return typeName(tv.T)
}

// applyObjectInfo copies the audit-relevant fields out of a gnolang
// ObjectInfo into the StateNode. Skips zero values so the template can
// branch on emptiness. Called by every decoder that handles a value type
// carrying ObjectInfo (StructValue, ArrayValue, MapValue, FuncValue,
// HeapItemValue, Block).
func applyObjectInfo(n *StateNode, info gno.ObjectInfo) {
	if info.ID != (gno.ObjectID{}) {
		n.ObjectID = info.ID.String()
	}
	if !info.Hash.IsZero() {
		n.Hash = hex.EncodeToString(info.Hash.Hashlet[:])
	}
	if info.OwnerID != (gno.ObjectID{}) {
		n.OwnerID = info.OwnerID.String()
	}
	if info.ModTime != 0 {
		n.ModTime = strconv.FormatUint(info.ModTime, 10)
	}
	if info.RefCount != 0 {
		n.RefCount = strconv.Itoa(info.RefCount)
	}
	if info.LastObjectSize != 0 {
		n.LastObjectSize = strconv.FormatInt(info.LastObjectSize, 10)
	}
}

// objectInfoOf extracts the ObjectInfo of an outer Value into a flat view
// shape suitable for sidebar display. Used by DecodeObjectFull to expose
// the queried object's metadata without re-parsing the JSON.
func objectInfoOf(v gno.Value) StateObjectInfoView {
	var info gno.ObjectInfo
	switch cv := v.(type) {
	case *gno.StructValue:
		info = cv.ObjectInfo
	case *gno.ArrayValue:
		info = cv.ObjectInfo
	case *gno.MapValue:
		info = cv.ObjectInfo
	case *gno.FuncValue:
		info = cv.ObjectInfo
	case *gno.HeapItemValue:
		info = cv.ObjectInfo
	case *gno.Block:
		info = cv.ObjectInfo
	default:
		return StateObjectInfoView{}
	}
	view := StateObjectInfoView{IsEscaped: info.IsEscaped}
	if !info.Hash.IsZero() {
		view.Hash = hex.EncodeToString(info.Hash.Hashlet[:])
	}
	if info.OwnerID != (gno.ObjectID{}) {
		view.OwnerID = info.OwnerID.String()
	}
	if info.ModTime != 0 {
		view.ModTime = strconv.FormatUint(info.ModTime, 10)
	}
	if info.RefCount != 0 {
		view.RefCount = strconv.Itoa(info.RefCount)
	}
	if info.LastObjectSize != 0 {
		view.LastObjectSize = strconv.FormatInt(info.LastObjectSize, 10)
	}
	return view
}

// DecodeObjectFull parses a vm/qobject_json (and optional vm/qtype_json)
// response and returns BOTH the children to render and the queried
// object's metadata. Single parse pass — handlers don't have to re-parse
// the JSON to surface ObjectInfo in the sidebar.
func DecodeObjectFull(rawObject, rawType []byte) (*DecodedObject, error) {
	var resp objectResponse
	if err := amino.UnmarshalJSON(rawObject, &resp); err != nil {
		return nil, fmt.Errorf("decode object JSON: %w", err)
	}

	info := objectInfoOf(resp.Value)

	var nodes []StateNode
	if len(rawType) == 0 {
		nodes = decodeValueChildren(resp.Value)
	} else {
		var typeResp struct {
			TypeID string   `json:"typeid"`
			Type   gno.Type `json:"type"`
		}
		if err := amino.UnmarshalJSON(rawType, &typeResp); err != nil || typeResp.Type == nil {
			nodes = decodeValueChildren(resp.Value)
		} else {
			nodes = decodeValueChildrenTyped(resp.Value, typeResp.Type, typeResp.TypeID)
		}
	}
	return &DecodedObject{Nodes: nodes, Info: info}, nil
}

// ---- Tiny helpers ----

func intPtr(n int) *int { return &n }
