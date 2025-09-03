package gnolang

import (
	"fmt"
	"slices"
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

func uintptrToBytes(u *uintptr) [sizeOfUintPtr]byte {
	return *(*[sizeOfUintPtr]byte)(unsafe.Pointer(u))
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

var ( // gno2: invar
	// These are extra special reserved keywords that cannot be used in gno code.
	// This set may be reduced in the future.
	_ = []string{"go", "gno", // XXX merge these. reservedNames2
		"invar",                                             // e.g. invar x []*Foo{}; cannot replace, but can modify elements.
		"undefined", "typed", "untyped", "typednil", "null", // types
		"across",     // related to cross
		"the", "THE", // definite article
		"pure", "perfect", "constant", // const/pure
		"beginning", "ending", "since", "till", "present", "current", "forever", "always", "never", // time & spatial (not included: past, history, future)
		"inside", "outside", "around", "through", "toward", "almost", "beyond", "across", "between", "where", "enter", "leave", "exit", // spatial
		"equals", "exactly", // equality
		"but", "despite", "except", "exception", "opposite", "than", "versus", "against", // difference
		"and", "nand", "or", "nor", "plus", "minus", "not", // binary (times?) & unary
		"set", "unset", "get", "put", "replace", "swap", "delete", "create", "construct", "destroy", // CRUD
		"map", "reduce", "join", "union", "constraint", "exists", "notexists", "unique", "nothing", "everything", "all", "any", "only", // set & operations & db
		"about", "and", "nor", "from", "amid", "into", "outof", "onto", "unto", "upon", "along", "via", "per", "regarding", "following", "within", "without", // relationship
		"would", "should", "could", "might", "must", "maybe", "definitely", "imagine", "pretend", // others
		"abort", "exit", "quit", "die", "kill", // execution
	}
)

// if true, caller should generally panic.
func isReservedName(n Name) bool {
	_, ok := reservedNames[n]
	return ok
}

// scans uverse static node for blocknames. (slow)
func isUverseName(n Name) bool {
	uverseNames := UverseNode().GetBlockNames()
	return slices.Contains(uverseNames, n)
}

//----------------------------------------
// other

// For keeping record of package & realm coins.
// If you need the bech32 address it is faster to call DerivePkgBech32Addr().
func DerivePkgCryptoAddr(pkgPath string) crypto.Address {
	if pkgPath == "" {
		panic("pkgpath cannot be empty")
	}
	b32addr, ok := IsGnoRunPath(pkgPath)
	if ok {
		addr, err := crypto.AddressFromBech32(b32addr)
		if err != nil {
			panic("invalid bech32 address in run path: " + pkgPath)
		}
		return addr
	}
	// NOTE: must not collide with pubkey addrs.
	return crypto.AddressFromPreimage([]byte("pkgPath:" + pkgPath))
}

func DerivePkgBech32Addr(pkgPath string) crypto.Bech32Address {
	if pkgPath == "" {
		panic("pkgpath cannot be empty")
	}
	b32addr, ok := IsGnoRunPath(pkgPath)
	if ok {
		return crypto.Bech32Address(b32addr)
	}
	// NOTE: must not collide with pubkey addrs.
	return crypto.AddressFromPreimage([]byte("pkgPath:" + pkgPath)).Bech32()
}

// Used to keep a record of storage deposit coins for a package or realm.
func DeriveStorageDepositCryptoAddr(pkgPath string) crypto.Address {
	if pkgPath == "" {
		panic("pkgpath cannot be empty in DeriveStorageDepositCryptoAddr()")
	}
	return crypto.AddressFromPreimage([]byte("pkgPath:" + pkgPath + ".storageDeposit"))
}

func DeriveStorageDepositBech32Addr(pkgPath string) crypto.Bech32Address {
	if pkgPath == "" {
		panic("pkgpath cannot be empty in DeriveStorageDepositBech32Addr()")
	}
	return crypto.AddressFromPreimage([]byte("pkgPath:" + pkgPath + ".storageDeposit")).Bech32()
}
