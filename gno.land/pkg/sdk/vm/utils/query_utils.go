package utils

import (
	"fmt"
	"math"
	"strconv"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// PkgResponse is the response of the "vm/qpkg_json" query.
type PkgResponse struct {
	Names  []string         `json:"names"`
	Values []gno.TypedValue `json:"values"`
}

// ObjectResponse is the response of the "vm/qobject_json" query.
type ObjectResponse struct {
	ObjectID string    `json:"objectid"`
	Value    gno.Value `json:"value"`
}

// TypeResponse is the response of the "vm/qtype_json" query.
type TypeResponse struct {
	TypeID string   `json:"typeid"`
	Type   gno.Type `json:"type"`
}

// EvalResponse is the response of the "vm/qeval_json" query.
type EvalResponse struct {
	Results []gno.TypedValue `json:"results"`
	Error   *string          `json:"@error,omitempty"`
}

// PrimitiveTypeName returns the name of the primitive type pt, for example
// "int" or "untyped bool".
func PrimitiveTypeName(pt gno.PrimitiveType) string {
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

// PrimitiveString formats the bool or numeric value in tv.N as the primitive
// type pt, for example "true", "-42" or "1.5". For other type, just return tv.String().
func PrimitiveString(pt gno.PrimitiveType, tv gno.TypedValue) string {
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

// GetTypeIDQueryString returns a TypeID for qtype_json lookup, drilling through
// PointerType so `*foo.Bar` resolves to Bar's TypeID.
func GetTypeIDQueryString(t gno.Type) string {
	if t == nil {
		return ""
	}
	switch tt := t.(type) {
	case gno.RefType:
		return tt.ID.String()
	case *gno.DeclaredType:
		// Mirror DeclaredTypeID: package/file decls are PkgPath.Name;
		// non-package-level decls disambiguate with [ParentLoc].
		if tt.ParentLoc.IsZero() {
			return tt.PkgPath + "." + string(tt.Name)
		}
		return tt.PkgPath + "[" + tt.ParentLoc.String() + "]." + string(tt.Name)
	case *gno.PointerType:
		return GetTypeIDQueryString(tt.Elt)
	}
	return ""
}

// ClampSliceWindow returns a safe [offset:end] window into a backing list of
// the given length, guarding against chain-supplied negative or out-of-range
// Offset/Length and against offset+length overflow.
func ClampSliceWindow(offset, length, listLen int) (int, int) {
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
