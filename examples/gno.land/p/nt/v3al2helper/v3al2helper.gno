// Package v3al2helper is a /p/-declared receiver type used to exercise
// the Layer-2 (storage-receiver) borrow path in v3a's walk. The type
// is intentionally trivial: a *Holder is created in one /r/, its method
// is called from another /r/, and the method reflects (Caller, Self) so
// the filetest can assert what the walk resolves to.
package v3al2helper

import "chain/runtime"

type Holder struct {
	Tag string
}

// Reflect returns (caller, self) from the holder's invocation context.
// Because Reflect is /p/-declared and the receiver is a *Holder stored
// in another /r/, the substrate's Layer-2 borrow shifts m.Realm to the
// holder's storage realm. The v3a walk should:
//   - report Self() as the storage realm (where the *Holder lives), and
//   - report Caller() as the realm that invoked Reflect.
func (h *Holder) Reflect() (caller runtime.Realm, self runtime.Realm) {
	return runtime.Caller(), runtime.Self()
}

// CallProbe invokes a probe function value from inside Holder's body.
// Because Holder.CallProbe is /p/-declared and the receiver is stored
// in another /r/, Layer-2 has already shifted m.Realm to the storage
// realm by the time fn is called. When fn is an /r/-declared
// function value, Layer-1 then shifts further to fn's realm. Used to
// exercise composed Layer-2 → Layer-1 walks.
func (h *Holder) CallProbe(fn func() (runtime.Realm, runtime.Realm, runtime.Realm)) (caller runtime.Realm, self runtime.Realm, prev runtime.Realm) {
	return fn()
}
