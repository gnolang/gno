package gnolang

import (
	"fmt"
	"strings"
	"unicode"
	"unsafe"

	"github.com/gnolang/gno/tm2/pkg/crypto"
)

//----------------------------------------
// Misc.

func cp(bz []byte) (ret []byte) {
	ret = make([]byte, len(bz))
	copy(ret, bz)
	return ret
}

// Returns the associated machine operation for binary AST operations.  TODO:
// to make this faster and inlineable, remove the switch statement and create a
// mathematical mapping between them.
func word2BinaryOp(w Word) Op {
	switch w {
	case ADD:
		return OpAdd
	case SUB:
		return OpSub
	case MUL:
		return OpMul
	case QUO:
		return OpQuo
	case REM:
		return OpRem
	case BAND:
		return OpBand
	case BOR:
		return OpBor
	case XOR:
		return OpXor
	case SHL:
		return OpShl
	case SHR:
		return OpShr
	case BAND_NOT:
		return OpBandn
	case LAND:
		return OpLand
	case LOR:
		return OpLor
	case EQL:
		return OpEql
	case LSS:
		return OpLss
	case GTR:
		return OpGtr
	case NEQ:
		return OpNeq
	case LEQ:
		return OpLeq
	case GEQ:
		return OpGeq
	default:
		panic(fmt.Sprintf("unexpected binary operation word %v", w.String()))
	}
}

func word2UnaryOp(w Word) Op {
	switch w {
	case ADD:
		return OpUpos
	case SUB:
		return OpUneg
	case NOT:
		return OpUnot
	case XOR:
		return OpUxor
	case MUL:
		panic("unexpected unary operation * - use StarExpr instead")
	case BAND:
		panic("unexpected unary operation & - use RefExpr instead")
	case ARROW:
		return OpUrecv
	default:
		panic("unexpected unary operation")
	}
}

func toString(n Node) string {
	if n == nil {
		return "<nil>"
	}
	return n.String()
}

// true if the first rune is uppercase.
func isUpper(s string) bool {
	var first rune
	for _, c := range s {
		first = c
		break
	}
	return unicode.IsUpper(first)
}

//----------------------------------------
// converting uintptr to bytes.

const sizeOfUintPtr = unsafe.Sizeof(uintptr(0))

func uintptrToBytes(u *uintptr) []byte {
	return (*[sizeOfUintPtr]byte)(unsafe.Pointer(u))[:]
}

func defaultPkgName(gopkgPath string) Name {
	parts := strings.Split(gopkgPath, "/")
	last := parts[len(parts)-1]
	parts = strings.Split(last, "-")
	name := parts[len(parts)-1]
	name = strings.ToLower(name)
	return Name(name)
}

//----------------------------------------
// value convenience

func toTypeValue(t Type) TypeValue {
	return TypeValue{
		Type: t,
	}
}

//----------------------------------------
// reserved & uverse names

var reservedNames = map[Name]struct{}{
	"break": {}, "default": {}, "func": {}, "interface": {}, "select": {},
	"case": {}, "defer": {}, "go": {}, "map": {}, "struct": {},
	"chan": {}, "else": {}, "goto": {}, "package": {}, "switch": {},
	"const": {}, "fallthrough": {}, "if": {}, "range": {}, "type": {},
	"continue": {}, "for": {}, "import": {}, "return": {}, "var": {},
}

// if true, caller should generally panic.
func isReservedName(n Name) bool {
	_, ok := reservedNames[n]
	return ok
}

// scans uverse static node for blocknames. (slow)
func isUverseName(n Name) bool {
	uverseNames := UverseNode().GetBlockNames()
	for _, name := range uverseNames {
		if name == n {
			return true
		}
	}
	return false
}

//----------------------------------------
// other

// For keeping record of package & realm coins.
func DerivePkgAddr(pkgPath string) crypto.Address {
	// NOTE: must not collide with pubkey addrs.
	return crypto.AddressFromPreimage([]byte("pkgPath:" + pkgPath))
}

// circular detection
// DeclNode represents a node in the dependency graph
// used to detect cycle definition of struct in `PredefineFileSet`.
type DeclNode struct {
	Name
	Loc          Location // file info
	Dependencies []*DeclNode
}

// insertDeclNode inserts a new dependency into the graph
func insertDeclNode(name Name, loc Location, deps ...Name) {
	var dep *DeclNode
	for _, d := range declGraph {
		if d.Name == name {
			dep = d
			dep.Loc = loc
			break
		}
	}
	if dep == nil {
		dep = &DeclNode{Name: name, Loc: loc}
		declGraph = append(declGraph, dep)
	}
	for _, depName := range deps {
		var child *DeclNode
		for _, d := range declGraph {
			if d.Name == depName {
				child = d
				break
			}
		}
		if child == nil {
			child = &DeclNode{Name: depName}
			declGraph = append(declGraph, child)
		}
		dep.Dependencies = append(dep.Dependencies, child)
	}
}

// assertNoCycle checks if there is a cycle in the declGraph graph
func assertNoCycle() {
	defer func() {
		declGraph = nil
	}()
	visited := make(map[Name]bool)
	reStack := make(map[Name]bool)
	var cycle []*DeclNode

	for _, dep := range declGraph {
		if detectCycle(dep, visited, reStack, &cycle) {
			cycleNames := make([]string, len(cycle))
			for i, c := range cycle {
				cycleNames[i] = fmt.Sprintf("%s(File: %s)", c.Name, c.Loc.File)
			}
			cycleMsg := strings.Join(cycleNames, " -> ")
			panic(fmt.Sprintf("Cyclic dependency detected: %s", cycleMsg))
		}
	}
}

// detectCycle detects cycle using DFS traversal
func detectCycle(node *DeclNode, visited, recStack map[Name]bool, cycle *[]*DeclNode) bool {
	if visited[node.Name] { // existing visited node are not in cycle, otherwise it wil be elided
		return false
	}
	visited[node.Name] = true
	recStack[node.Name] = true
	*cycle = append(*cycle, node)

	for _, d := range node.Dependencies {
		// check if d is in recStack to form a cycle
		if recStack[d.Name] {
			for _, n := range *cycle {
				if n == d {
					startIndex := 0
					for ; (*cycle)[startIndex] != d; startIndex++ {
					}
					*cycle = append((*cycle)[startIndex:], d)
				}
			}
			return true
		} else {
			if detectCycle(d, visited, recStack, cycle) {
				return true
			}
		}
	}

	delete(recStack, node.Name)
	// Backtrack: Remove the last node from the cycle slice and mark as not in recStack
	*cycle = (*cycle)[:len(*cycle)-1]

	return false
}

func checkFieldReference(PkgPath string, t Type, names *[]Name) bool {
	switch fdt := t.(type) {
	case *DeclaredType:
		if PkgPath == fdt.PkgPath { // not cross pkg
			*names = append(*names, fdt.Name)
			return true
		}
	case *ArrayType:
		return checkFieldReference(PkgPath, fdt.Elem(), names)
	default:
		return false
	}

	return false
}
