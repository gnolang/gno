package gnolang

// The texts of the write rule's refusals, each `<what was refused> (<do this
// instead>)`; refusals_test.go checks every refusal site uses one of these.

const (
	// /p/ code writes an object another realm owns; then the target expression.
	refuseLibWrite = "cannot modify object owned by another realm from library code (call a function of the owning realm): "
	// After the realm path: a function value called by value with no live loan writes.
	refuseByValue = "'s code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): "
	// After the realm path: the realm's own code, entered without crossing, writes.
	refuseUncrossed = "'s code was entered without crossing and is read-only (call its crossing function): "
	// Any other write to an object stored in another realm.
	refuseTainted = "cannot directly modify readonly tainted object (call a crossing function of the owning realm): "
	// At save: an object a read-only stretch wrote would be stored.
	refuseStoreUncrossed = "%s: an object written by read-only code would be stored (write it inside a crossing call of %s)"
	// A plain function is called with cross.
	refuseNotCrossing = "cannot cross-call a non-crossing function %s.%v from %s"
	// Another realm's crossing function, reached by value, is called with cur.
	refuseExternalCur = "cannot cur-call to external realm function %s.%v from %s"
	// append, copy or delete writes into a read-only slice or map.
	refuseAppend = "cannot append to readonly tainted slice"
	refuseCopy   = "cannot copy to readonly tainted slice"
	refuseDelete = "cannot delete from readonly tainted map"
	// Another realm's stored value is converted to another mutable type.
	refuseConvStored = "illegal conversion of readonly or externally stored value"
	// A value is converted to a type another realm declares.
	refuseConvForeign = "illegal conversion to external realm type"
	// A /p/ package's own state is written after its init.
	refuseImmutablePkg = "cannot mutate %s: package is immutable post-init"
)

// refusalTexts lists every constant above, for the tests and the docs table.
var refusalTexts = []string{
	refuseLibWrite, refuseByValue, refuseUncrossed, refuseTainted,
	refuseStoreUncrossed, refuseNotCrossing, refuseExternalCur,
	refuseAppend, refuseCopy, refuseDelete,
	refuseConvStored, refuseConvForeign, refuseImmutablePkg,
}

// refuseWrite raises the Gno panic that refuses a write to lx, naming why.
func (m *Machine) refuseWrite(lx Expr) {
	m.Panic(typedString(m.readonlyAccessPanic(lx)))
}
