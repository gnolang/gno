package gnolang

// The texts of the write rule's refusals, each `<what was refused> (<do this
// instead>)`; refusals_test.go checks every refusal site uses one of these.

const (
	// /p/ code writes an object another realm owns; then the target expression.
	refuseLibWrite = "cannot modify object owned by another realm from library code (call a crossing function of the owning realm): "
	// After the realm path: a function value called by value with no live loan writes.
	refuseByValue = "'s code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): "
	// After the realm path: the realm's own code, entered without crossing, writes.
	refuseUncrossed = "'s code was entered without crossing and is read-only (call its crossing function): "
	// Any other refused write: another realm's or a /p/ package's object, or
	// the realm's own object inside code that runs view-only.
	refuseTainted = "cannot directly modify readonly tainted object (write it in a named function of the realm that owns it, during a crossing call into that realm): "
	// At save: an object a read-only stretch wrote would be stored.
	refuseStoreUncrossed = "%s: an object written by read-only code would be stored (make the function that wrote it a crossing function of %s and call it with cross)"
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

// refusal is one refusal of the write rule: its text, when it fires and what
// to do; master marks a text kept byte-identical to master's.
type refusal struct {
	text, when, fix string
	master          bool
}

// refusals lists every constant above; the docs table is generated from it.
var refusals = []refusal{
	{refuseLibWrite,
		"/p/ code writes an object another realm owns",
		"Call a crossing function of the realm that owns the object; a plain function of it is refused too.",
		false},
	{refuseByValue,
		"a function value called by value with no live loan writes",
		"Give it a cur realm parameter and call it with cross.",
		false},
	{refuseUncrossed,
		"a realm's own code, entered without crossing, writes",
		"Call its crossing function.",
		false},
	{refuseTainted,
		"any other refused write: to another realm's or a /p/ package's object, or to your realm's own object in code that runs view-only",
		"Another realm's object: call a crossing function of that realm. Your realm's own object, made read-only because a function value or a library you passed it to does the write: do the write in your own named function, pass the callback in place, or import the writing method's package directly. A /p/ package's object cannot be written after its init: keep state that changes in a realm.",
		false},
	{refuseStoreUncrossed,
		"at save, an object that read-only code wrote would be stored",
		"Give the function that wrote it a cur realm parameter in the realm that stores it, and call it with cross.",
		false},
	{refuseNotCrossing,
		"a plain function is called with cross",
		"Call it without cross.",
		true},
	{refuseExternalCur,
		"another realm's crossing function, reached through a function value, is called with cur",
		"The author of the code named after \"from\" changes the call to cross(cur); the sender of the transaction cannot. If that code is another realm's, use a function of it that crosses.",
		true},
	{refuseAppend,
		"append writes into a slice the running code may not write",
		"Another realm's slice: call a crossing function of that realm. Your realm's own slice, made read-only because a function value or a library you passed it to does the write: do the write in your own named function, pass the callback in place, or import the writing method's package directly.",
		true},
	{refuseCopy,
		"copy writes into a slice the running code may not write",
		"Another realm's slice: call a crossing function of that realm. Your realm's own slice, made read-only because a function value or a library you passed it to does the write: do the write in your own named function, pass the callback in place, or import the writing method's package directly.",
		true},
	{refuseDelete,
		"delete writes into a map the running code may not write",
		"Another realm's map: call a crossing function of that realm. Your realm's own map, made read-only because a function value or a library you passed it to does the write: do the write in your own named function, pass the callback in place, or import the writing method's package directly.",
		true},
	{refuseConvStored,
		"another realm's stored value is converted to another mutable type",
		"Convert it inside a crossing function of the realm that stores it.",
		true},
	{refuseConvForeign,
		"a value is converted to a type another realm declares",
		"Call a function of the realm that declares the type to make it.",
		true},
	{refuseImmutablePkg,
		"a /p/ package's own state is written after its init",
		"Keep state that changes in a realm.",
		true},
}

// refuseWrite raises the Gno panic that refuses a write to lx, naming why.
func (m *Machine) refuseWrite(lx Expr) {
	m.Panic(typedString(m.readonlyAccessPanic(lx)))
}
