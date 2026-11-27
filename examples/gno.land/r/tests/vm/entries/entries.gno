// Package entries is a fixture for entry-function authority: stored objects
// are written only by code in this realm's source, never by foreign callers
// through references, promoted methods, interfaces or method values.
package entries

import (
	"gno.land/p/demo/tests/viewgrant/v0"
	"gno.land/p/nt/avl/v0"
)

var tree = avl.NewTree()

func init() {
	tree.Set("k", "v0")
	registry = append(registry, Inc)
}

func read(t *avl.Tree, k string) string {
	if v := t.Get(k); v != nil {
		return v.(string)
	}
	return ""
}

func GetTree() *avl.Tree { return tree }

func Read(k string) string { return read(tree, k) }

func OwnerSet(cur realm, k, v string) { tree.Set(k, v) }

// OwnerSetDeferred writes from a defer in this realm's own crossing entry.
func OwnerSetDeferred(cur realm, k, v string) { defer tree.Set(k, v) }

// Holder embeds a /p/ tree, so Set is a promoted library method.
type Holder struct{ *avl.Tree }

var holder = &Holder{avl.NewTree()}

func GetHolder() *Holder { return holder }

func ReadHolder(k string) string { return read(holder.Tree, k) }

func OwnerSetHolder(cur realm, k, v string) { holder.Set(k, v) }

// Setter exposes the stored tree behind an interface.
type Setter interface{ Set(k string, v any) bool }

var ifaceTree = avl.NewTree()

func AsSetter() Setter { return ifaceTree }

func ReadIface(k string) string { return read(ifaceTree, k) }

func OwnerSetIface(cur realm, k, v string) { ifaceTree.Set(k, v) }

// NewTree allocates a fresh tree owned by this realm.
func NewTree() *avl.Tree { return avl.NewTree() }

func ReadFrom(t *avl.Tree, k string) string { return read(t, k) }

func OwnerSetTree(cur realm, t *avl.Tree, k, v string) { t.Set(k, v) }

// b is the victim tree for an argument write.
var b = avl.NewTree()

func GetB() *avl.Tree { return b }

// pair is this realm's own Pair; its CopyTo writes its argument.
var pair = viewgrant.NewPair()

func GetPair() *viewgrant.Pair { return pair }

func ReadB() string { return read(b, "b") }

// ReadBFrom reads the key an argument-writing /p/ method writes, so a probe
// sees the write even when it lands.
func ReadBFrom() string { return read(b, "from") }

func OwnerSetB(cur realm, v string) { b.Set("b", v) }

// registry stores a crossing function written in this realm's source.
var (
	registry []func(realm)
	counter  int
)

func Entry(i int) func(realm) { return registry[i] }

func Inc(cur realm) { counter++ }

func Count() int { return counter }
