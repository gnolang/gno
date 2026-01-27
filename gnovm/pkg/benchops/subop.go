package benchops

// SubOp represents a sub-operation within an opcode.
// Sub-operations provide finer granularity for profiling,
// tracking individual variable assignments within statements like "i, j = 1, 2".
type SubOp byte

// SubOp constants for sub-operation tracing.
// These represent the specific sub-operations that can occur within an opcode.
const (
	SubOpNone SubOp = 0x00 // No sub-operation

	// Assignment sub-operations (0x01-0x0F)
	SubOpAssignVar   SubOp = 0x01 // a = v (simple variable assignment)
	SubOpDefineVar   SubOp = 0x02 // a := v (variable definition)
	SubOpAssignIndex SubOp = 0x03 // a[i] = v (index assignment)
	SubOpAssignField SubOp = 0x04 // s.f = v (field assignment)

	// Range sub-operations (0x10-0x1F)
	SubOpRangeKey   SubOp = 0x10 // range key assignment
	SubOpRangeValue SubOp = 0x11 // range value assignment

	// Composite literal sub-operations (0x20-0x2F)
	SubOpArrayElem   SubOp = 0x20 // array literal element
	SubOpSliceElem   SubOp = 0x21 // slice literal element (reserved, not yet instrumented)
	SubOpMapEntry    SubOp = 0x22 // map literal entry
	SubOpStructField SubOp = 0x23 // struct literal field

	// Function sub-operations (0x30-0x3F)
	SubOpParamAssign SubOp = 0x30 // function parameter assignment
	SubOpResultCopy  SubOp = 0x31 // return result copy

	maxSubOps = 256
)

// String returns the name of the sub-operation.
func (s SubOp) String() string {
	switch s {
	case SubOpNone:
		return "None"
	case SubOpAssignVar:
		return "AssignVar"
	case SubOpDefineVar:
		return "DefineVar"
	case SubOpAssignIndex:
		return "AssignIndex"
	case SubOpAssignField:
		return "AssignField"
	case SubOpRangeKey:
		return "RangeKey"
	case SubOpRangeValue:
		return "RangeValue"
	case SubOpArrayElem:
		return "ArrayElem"
	case SubOpSliceElem:
		return "SliceElem"
	case SubOpMapEntry:
		return "MapEntry"
	case SubOpStructField:
		return "StructField"
	case SubOpParamAssign:
		return "ParamAssign"
	case SubOpResultCopy:
		return "ResultCopy"
	default:
		return "Unknown"
	}
}

// SubOpContext provides context for a sub-operation measurement.
type SubOpContext struct {
	File    string // Source file name
	Line    int    // Line number in source
	VarName string // Variable/field name being assigned (mutually exclusive with Index)
	Index   int    // Index for indexed operations (0-based; only meaningful when VarName is empty)
}
