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

// Kind constants enumerate the shapes the walker emits. Kept as untyped
// strings so html/template's `eq` comparator works against `.Kind`.
const (
	KindPrimitive = "primitive"
	KindStruct    = "struct"
	KindArray     = "array"
	KindSlice     = "slice"
	KindMap       = "map"
	KindPointer   = "pointer"
	KindRef       = "ref"
	KindFunc      = "func"
	KindClosure   = "closure"
	KindType      = "type"
	KindInterface = "interface"
	KindPackage   = "package"
	KindNil       = "nil"
	KindCycle     = "cycle"
	KindTruncated = "truncated"
)

// StateNode is the UI-friendly decoded representation of a gno value.
// Built by the walker from raw Amino JSON; enriched post-walk with Href,
// SourceHTML, Doc, and Anchor by the orchestrator and sidebar builders.
type StateNode struct {
	Name       string
	Type       string
	Kind       string
	Value      string
	Expandable bool
	Children   []StateNode
	ObjectID   string
	TypeID     string
	Length     *int
	// Preview is a one-line summary of Children, rendered in
	// collapsed/ref rows. Re-computed after lazy ref fetches.
	Preview string
	Source  *SourceLocation
	// SourceHTML carries chroma-highlighted code; template.HTML so
	// html/template trusts it as already-safe markup.
	SourceHTML template.HTML
	// Href / OwnerHref are typed template.URL so html/template trusts them.
	Href      template.URL
	OwnerHref template.URL
	// Anchor is the row id stamped by Build{Package,Object}Sidebar for #
	// fragment linking from the TOC.
	Anchor string
	// ObjectInfo metadata captured by the walker from qobject_json/qpkg_json.
	Hash           string
	OwnerID        string
	ModTime        string
	RefCount       string
	LastObjectSize string
	// Doc is the plain-text source comment attached post-walk from the
	// package's JSON doc index, matched by Name. Rendered text-escaped
	// by the template — no Markdown processing.
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

// DecodeObjectJSON decodes a vm/qobject_json response into the contained
// object's children. Struct fields fall back to positional indices without
// a type context — use DecodeObjectJSONWithType for field names.
func DecodeObjectJSON(raw []byte) ([]StateNode, error) {
	var resp objectResponse
	if err := amino.UnmarshalJSON(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode object JSON: %w", err)
	}
	return decodeValueChildren(resp.Value), nil
}

// DecodeObjectJSONWithType decodes a vm/qobject_json response together
// with a vm/qtype_json response so struct field names replace positional
// indices. Nil/empty rawType falls back to plain DecodeObjectJSON.
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

// decodeValueChildrenTyped is decodeValueChildren plus an outer Type
// used to resolve struct field names; originalTid is forwarded into
// nested ref nodes so subsequent fetch rounds still resolve names.
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
			total := len(sv.Fields)
			shown := total
			if shown > maxChildrenPerNode {
				shown = maxChildrenPerNode
			}
			children := make([]StateNode, shown, shown+1)
			for i := 0; i < shown; i++ {
				name := strconv.Itoa(i)
				if i < len(st.Fields) && st.Fields[i].Name != "" {
					name = string(st.Fields[i].Name)
				}
				children[i] = decodeTypedValue(name, sv.Fields[i])
			}
			if total > shown {
				children = append(children, truncatedChildrenNode(total-shown))
			}
			return children
		}
	}

	// Other shapes: use the existing un-typed children logic.
	return decodeValueChildren(v)
}

// ---- Core walker ----

// Walker bounds — defenses against pathological / hostile values.
// maxDecodeDepth stops stack-overflow recursion; maxChildrenPerNode
// bounds DOM size for giant collections (surplus collapses to one
// truncated sentinel).
const (
	maxDecodeDepth     = 256
	maxChildrenPerNode = 500
)

// tooDeepNode is the sentinel emitted when a subtree exceeds maxDecodeDepth.
func tooDeepNode(name string) StateNode {
	return StateNode{Name: name, Type: "(too deep)", Kind: KindTruncated, Value: "…"}
}

// truncatedChildrenNode summarises entries dropped past maxChildrenPerNode.
func truncatedChildrenNode(remaining int) StateNode {
	return StateNode{
		Name: "…", Kind: KindTruncated,
		Value: fmt.Sprintf("(%d more entries omitted)", remaining),
	}
}

// clampSliceWindow returns a safe [offset:end] window into a backing list of
// the given length, guarding against chain-supplied negative or out-of-range
// Offset/Length and against offset+length overflow.
func clampSliceWindow(offset, length, listLen int) (int, int) {
	if offset < 0 {
		offset = 0
	}
	if offset > listLen {
		offset = listLen
	}
	if length < 0 {
		length = 0
	}
	end := offset + length
	if end < offset || end > listLen { // overflow or past end
		end = listLen
	}
	return offset, end
}

func decodeTypedValue(name string, tv gno.TypedValue) StateNode {
	return decodeTypedValueAt(0, name, tv)
}

func decodeTypedValueAt(depth int, name string, tv gno.TypedValue) StateNode {
	if depth >= maxDecodeDepth {
		return tooDeepNode(name)
	}
	if tv.T == nil {
		return StateNode{Name: name, Type: "<nil>", Kind: KindNil, Value: "nil"}
	}

	tName := typeName(tv.T)
	kind := typeKind(tv.T)
	bt := baseType(tv.T)
	typeID := getTypeID(tv.T)

	// FuncType stored as RefValue → expandable to fetch source on detail page.
	if _, isFunc := bt.(*gno.FuncType); isFunc {
		if rv, ok := tv.V.(gno.RefValue); ok {
			return StateNode{
				Name: name, Type: funcSignature(tv.T), Kind: KindFunc,
				Expandable: true, ObjectID: rv.ObjectID.String(),
			}
		}
	}

	// RefValue: persisted object reference.
	if rv, ok := tv.V.(gno.RefValue); ok {
		if rv.PkgPath != "" {
			return StateNode{Name: name, Type: tName, Kind: KindPackage, Value: rv.PkgPath}
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
		return StateNode{Name: name, Type: "type", Kind: KindType, Value: typeName(tv2.Type)}
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

// decodeValueChildren turns a queried Object's Value into the children
// to display. Collection-shaped values yield one StateNode per element;
// scalar shapes yield a single representative node so the page is never
// empty.
func decodeValueChildren(v gno.Value) []StateNode {
	switch cv := v.(type) {
	case nil:
		return nil

	// ---- Collection-shaped: render fields/elements as direct children ----

	case *gno.StructValue:
		total := len(cv.Fields)
		shown := total
		if shown > maxChildrenPerNode {
			shown = maxChildrenPerNode
		}
		nodes := make([]StateNode, shown, shown+1)
		for i := 0; i < shown; i++ {
			nodes[i] = decodeTypedValue(strconv.Itoa(i), cv.Fields[i])
		}
		if total > shown {
			nodes = append(nodes, truncatedChildrenNode(total-shown))
		}
		return nodes
	case *gno.ArrayValue:
		if cv.Data != nil {
			n := len(cv.Data)
			return []StateNode{{
				Name: "data", Type: "[]byte", Kind: KindPrimitive,
				Value: fmt.Sprintf("[%d]byte{...}", n),
			}}
		}
		total := len(cv.List)
		shown := total
		if shown > maxChildrenPerNode {
			shown = maxChildrenPerNode
		}
		nodes := make([]StateNode, shown, shown+1)
		for i := 0; i < shown; i++ {
			nodes[i] = decodeTypedValue(strconv.Itoa(i), cv.List[i])
		}
		if total > shown {
			nodes = append(nodes, truncatedChildrenNode(total-shown))
		}
		return nodes
	case *gno.SliceValue:
		// Slices live in their backing array; if the array is inline, expose
		// the visible window as elements. If the array is itself a stored
		// ref, expose a single ref node so the user can navigate.
		if av, ok := cv.Base.(*gno.ArrayValue); ok && av.Data == nil {
			offset, end := clampSliceWindow(cv.Offset, cv.Length, len(av.List))
			window := av.List[offset:end]
			total := len(window)
			shown := total
			if shown > maxChildrenPerNode {
				shown = maxChildrenPerNode
			}
			nodes := make([]StateNode, shown, shown+1)
			for i := 0; i < shown; i++ {
				nodes[i] = decodeTypedValue(strconv.Itoa(i), window[i])
			}
			if total > shown {
				nodes = append(nodes, truncatedChildrenNode(total-shown))
			}
			return nodes
		}
		if rv, ok := cv.Base.(gno.RefValue); ok {
			n := cv.Length
			return []StateNode{{
				Name: "(slice)", Type: "[]…", Kind: KindSlice,
				Expandable: n > 0, ObjectID: rv.ObjectID.String(), Length: intPtr(n),
			}}
		}
		return nil
	case *gno.MapValue:
		var (
			nodes []StateNode
			total int
		)
		if cv.List != nil {
			for cur := cv.List.Head; cur != nil; cur = cur.Next {
				total++
				if len(nodes) >= maxChildrenPerNode {
					continue
				}
				keyStr := previewTypedValue(cur.Key)
				nodes = append(nodes, decodeTypedValue(keyStr, cur.Value))
			}
		}
		if total > len(nodes) {
			nodes = append(nodes, truncatedChildrenNode(total-len(nodes)))
		}
		return nodes
	case *gno.HeapItemValue:
		// Unwrap: when the inner is a collection, return its children
		// directly (no redundant "value :" row); scalars stay labelled.
		inner := decodeTypedValue("value", cv.Value)
		if len(inner.Children) > 0 {
			return inner.Children
		}
		return []StateNode{inner}
	case *gno.Block:
		total := len(cv.Values)
		shown := total
		if shown > maxChildrenPerNode {
			shown = maxChildrenPerNode
		}
		nodes := make([]StateNode, shown, shown+1)
		for i := 0; i < shown; i++ {
			nodes[i] = decodeTypedValue(strconv.Itoa(i), cv.Values[i])
		}
		if total > shown {
			nodes = append(nodes, truncatedChildrenNode(total-shown))
		}
		return nodes

	// ---- Scalar-shaped: wrap into a single representative node ----

	case *gno.FuncValue:
		name := string(cv.Name)
		if name == "" {
			name = "(function)"
		}
		return []StateNode{decodeFuncInline(0, name, cv)}
	case *gno.BoundMethodValue:
		return []StateNode{decodeFuncInline(0, "(method)", cv.Func)}
	case gno.PointerValue:
		if cv.TV != nil {
			return []StateNode{decodeTypedValue("*", *cv.TV)}
		}
		if rv, ok := cv.Base.(gno.RefValue); ok {
			return []StateNode{{
				Name: "*", Type: "(stored)", Kind: KindPointer,
				Expandable: true, ObjectID: rv.ObjectID.String(),
			}}
		}
		return []StateNode{{Name: "*", Type: "*", Kind: KindPointer, Value: "nil"}}
	case gno.TypeValue:
		return []StateNode{{
			Name: "(type)", Type: "type", Kind: KindType,
			Value: typeName(cv.Type),
		}}
	case gno.RefValue:
		return []StateNode{{
			Name: "(ref)", Type: "(stored)", Kind: KindRef,
			Expandable: true, ObjectID: cv.ObjectID.String(),
		}}
	case gno.ExportRefValue:
		return []StateNode{{
			Name: "(cycle)", Type: "(cycle)", Kind: KindCycle,
			Value: fmt.Sprintf("<cycle %s>", cv.ObjectID),
		}}
	case gno.StringValue:
		return []StateNode{{
			Name: "(string)", Type: "string", Kind: KindPrimitive,
			Value: quoteString(string(cv)),
		}}
	}
	return nil
}

// ---- Per-kind decoders ----

func decodePrimitive(name, tName string, pt gno.PrimitiveType, tv gno.TypedValue) StateNode {
	// Strings live in V (StringValue); all other primitives in N.
	if pt == gno.StringType || pt == gno.UntypedStringType {
		s := tv.GetString()
		return StateNode{Name: name, Type: tName, Kind: KindPrimitive, Value: quoteString(s)}
	}
	if tv.V == nil && tv.N == [8]byte{} {
		return StateNode{Name: name, Type: tName, Kind: KindPrimitive, Value: zeroValueFor(pt)}
	}
	return StateNode{Name: name, Type: tName, Kind: KindPrimitive, Value: primitiveDisplay(pt, tv)}
}

func decodeStruct(depth int, name, tName, typeID string, bt gno.Type, sv *gno.StructValue) StateNode {
	fieldNames := structFieldNames(bt)
	total := len(sv.Fields)
	shown := total
	if shown > maxChildrenPerNode {
		shown = maxChildrenPerNode
	}
	children := make([]StateNode, shown, shown+1)
	for i := 0; i < shown; i++ {
		var fname string
		if i < len(fieldNames) && fieldNames[i] != "" {
			fname = fieldNames[i]
		} else {
			fname = strconv.Itoa(i)
		}
		children[i] = decodeTypedValueAt(depth+1, fname, sv.Fields[i])
	}
	if total > shown {
		children = append(children, truncatedChildrenNode(total-shown))
	}
	node := StateNode{
		Name: name, Type: tName, Kind: KindStruct,
		Expandable: total > 0, Children: children,
		TypeID:  typeID,
		Length:  intPtr(total),
		Preview: buildChildrenPreview(children),
	}
	applyObjectInfo(&node, sv.ObjectInfo)
	return node
}

func decodeArray(depth int, name, tName string, av *gno.ArrayValue) StateNode {
	if av.Data != nil {
		n := len(av.Data)
		node := StateNode{
			Name: name, Type: tName, Kind: KindArray,
			Value:  fmt.Sprintf("[%d]byte{...}", n),
			Length: intPtr(n),
		}
		applyObjectInfo(&node, av.ObjectInfo)
		return node
	}
	total := len(av.List)
	visible := total
	if visible > maxChildrenPerNode {
		visible = maxChildrenPerNode
	}
	children := make([]StateNode, visible, visible+1)
	for i := 0; i < visible; i++ {
		children[i] = decodeTypedValueAt(depth+1, strconv.Itoa(i), av.List[i])
	}
	if total > visible {
		children = append(children, truncatedChildrenNode(total-visible))
	}
	node := StateNode{
		Name: name, Type: tName, Kind: KindArray,
		Expandable: len(children) > 0, Children: children,
		Length: intPtr(total),
	}
	applyObjectInfo(&node, av.ObjectInfo)
	return node
}

func decodeSlice(depth int, name, tName string, sv *gno.SliceValue) StateNode {
	length := sv.Length
	if rv, ok := sv.Base.(gno.RefValue); ok {
		return StateNode{
			Name: name, Type: tName, Kind: KindSlice,
			Expandable: length > 0, ObjectID: rv.ObjectID.String(), Length: intPtr(length),
		}
	}
	if av, ok := sv.Base.(*gno.ArrayValue); ok {
		if av.Data != nil {
			return StateNode{
				Name: name, Type: tName, Kind: KindSlice,
				Value:  fmt.Sprintf("[]byte (len=%d)", length),
				Length: intPtr(length),
			}
		}
		offset, end := clampSliceWindow(sv.Offset, length, len(av.List))
		window := av.List[offset:end]
		total := len(window)
		shown := total
		if shown > maxChildrenPerNode {
			shown = maxChildrenPerNode
		}
		children := make([]StateNode, shown, shown+1)
		for i := 0; i < shown; i++ {
			children[i] = decodeTypedValueAt(depth+1, strconv.Itoa(i), window[i])
		}
		if total > shown {
			children = append(children, truncatedChildrenNode(total-shown))
		}
		return StateNode{
			Name: name, Type: tName, Kind: KindSlice,
			Expandable: len(children) > 0, Children: children,
			Length: intPtr(length),
		}
	}
	return StateNode{
		Name: name, Type: tName, Kind: KindSlice,
		Expandable: length > 0, Length: intPtr(length),
	}
}

func decodeMap(depth int, name, tName string, mv *gno.MapValue) StateNode {
	var (
		children []StateNode
		total    int
	)
	if mv.List != nil {
		for cur := mv.List.Head; cur != nil; cur = cur.Next {
			total++
			if len(children) >= maxChildrenPerNode {
				continue
			}
			keyStr := previewTypedValue(cur.Key)
			children = append(children, decodeTypedValueAt(depth+1, keyStr, cur.Value))
		}
	}
	if total > len(children) {
		children = append(children, truncatedChildrenNode(total-len(children)))
	}
	node := StateNode{
		Name: name, Type: tName, Kind: KindMap,
		Expandable: len(children) > 0, Children: children,
		Length:  intPtr(total),
		Preview: buildChildrenPreview(children),
	}
	applyObjectInfo(&node, mv.ObjectInfo)
	return node
}

func decodePointer(depth int, name, tName, typeID string, pv gno.PointerValue) StateNode {
	if rv, ok := pv.Base.(gno.RefValue); ok {
		return StateNode{
			Name: name, Type: tName, Kind: KindPointer,
			Expandable: true, ObjectID: rv.ObjectID.String(),
			// Carry pointee TypeID so qtype_json can resolve struct field names.
			TypeID: typeID,
		}
	}
	if pv.TV != nil {
		child := decodeTypedValueAt(depth+1, "*", *pv.TV)
		return StateNode{
			Name: name, Type: tName, Kind: KindPointer,
			Expandable: true, Children: []StateNode{child},
		}
	}
	return StateNode{Name: name, Type: tName, Kind: KindPointer, Value: "nil"}
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
	kind := KindFunc
	if hasCaps {
		kind = KindClosure
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
	// Expandable only when Source is available to disclose the body.
	return StateNode{
		Name: name, Type: sig, Kind: kind,
		Expandable: src != nil, Source: src,
	}
}

// ---- Type helpers ----

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
		return KindNil
	}
	switch tt := t.(type) {
	case gno.PrimitiveType:
		return KindPrimitive
	case *gno.PointerType:
		return KindPointer
	case *gno.ArrayType:
		return KindArray
	case *gno.SliceType:
		return KindSlice
	case *gno.StructType:
		return KindStruct
	case *gno.MapType:
		return KindMap
	case *gno.FuncType:
		return KindFunc
	case *gno.InterfaceType:
		return KindInterface
	case gno.RefType:
		return KindRef
	case *gno.DeclaredType:
		return typeKind(tt.Base)
	case *gno.TypeType:
		return KindType
	case *gno.PackageType:
		return KindPackage
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

// structFieldNames returns field names for a StructType (unwrapping DeclaredType).
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

// getTypeID returns a TypeID for qtype_json lookup, drilling through
// PointerType so `*foo.Bar` resolves to Bar's TypeID.
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

// funcSignature builds a human-readable signature, hiding the implicit
// `cur realm` crossing parameter.
func funcSignature(t gno.Type) string {
	ft, ok := t.(*gno.FuncType)
	if !ok {
		return "func()"
	}
	params := make([]string, 0, len(ft.Params))
	for _, p := range ft.Params {
		// Hide implicit crossing param: cur-prefixed RefType.
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

const inlinePreviewMaxFields = 3

// buildChildrenPreview formats children as `{name: "alice", age: 30, …}`
// for the collapsed row preview. Returns "" when empty.
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

// previewChildValue returns a compact string for the inline preview:
// leaf value if present, else type (with length when known).
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

// applyObjectInfo copies audit-relevant ObjectInfo fields onto the node,
// skipping zero values so the template can branch on emptiness.
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

// objectInfoOf extracts ObjectInfo from a Value into the sidebar view shape.
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

// DecodeObjectFull parses qobject_json (and optional qtype_json) into
// children to render plus the queried object's ObjectInfo in one pass.
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
