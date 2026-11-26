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
	refuseNotCrossing = "cannot cross-call a non-crossing function %s.%v from %s (call it without cross)"
	// Another realm's crossing function, reached by value, is called with cur.
	refuseExternalCur = "cannot cur-call to external realm function %s.%v from %s (call it with cross(cur))"
	// append, copy or delete writes into a read-only slice or map.
	refuseAppend = "cannot append to readonly tainted slice (call a crossing function of the owning realm)"
	refuseCopy   = "cannot copy to readonly tainted slice (call a crossing function of the owning realm)"
	refuseDelete = "cannot delete from readonly tainted map (call a crossing function of the owning realm)"
	// Another realm's stored value is converted to another mutable type.
	refuseConvStored = "illegal conversion of readonly or externally stored value (convert it inside a crossing function of the realm that stores it)"
	// A value is converted to a type another realm declares.
	refuseConvForeign = "illegal conversion to external realm type (call a function of the realm that declares the type to make it)"
	// A /p/ package's own state is written after its init.
	refuseImmutablePkg = "cannot mutate %s: package is immutable post-init (keep state that changes in a realm)"
)

// refusal is one refusal of the write rule: its text, when it fires and what
// to do; master marks a text kept byte-identical to master's.
type refusal struct {
	text, when, fix string
	master          bool
}

// refusals lists every constant above; the docs table is generated from it.
var refusals = []refusal{
	{refuseLibWrite, "/p/ code writes an object another realm owns", "Call a function of the owning realm.", false},
	{refuseByValue, "a function value called by value with no live loan writes", "Give it a cur realm parameter and call it with cross.", false},
	{refuseUncrossed, "a realm's own code, entered without crossing, writes", "Call its crossing function.", false},
	{refuseTainted, "any other write to an object stored in another realm", "Call a crossing function of the owning realm.", false},
	{refuseStoreUncrossed, "at save, an object read-only code wrote would be stored", "Write it inside a crossing call of the realm.", false},
	{refuseNotCrossing, "a plain function is called with cross", "Call it without cross.", false},
	{refuseExternalCur, "another realm's crossing function, reached by value, is called with cur", "Call it with cross(cur).", false},
	{refuseAppend, "append writes into a read-only slice", "Call a crossing function of the owning realm.", false},
	{refuseCopy, "copy writes into a read-only slice", "Call a crossing function of the owning realm.", false},
	{refuseDelete, "delete writes into a read-only map", "Call a crossing function of the owning realm.", false},
	{refuseConvStored, "another realm's stored value is converted to another mutable type", "Convert it inside a crossing function of the realm that stores it.", false},
	{refuseConvForeign, "a value is converted to a type another realm declares", "Call a function of the realm that declares the type to make it.", false},
	{refuseImmutablePkg, "a /p/ package's own state is written after its init", "Keep state that changes in a realm.", false},
}

// refuseWrite raises the Gno panic that refuses a write to lx, naming why.
func (m *Machine) refuseWrite(lx Expr) {
	m.Panic(typedString(m.readonlyAccessPanic(lx)))
}
