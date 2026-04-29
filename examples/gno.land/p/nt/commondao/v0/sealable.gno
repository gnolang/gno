package commondao

// Sealable defines an interface for making types readonly.
// TODO: Improve when generics are supported by Gno
type Sealable interface {
	// Seal returns a sealed copy of the sealable type.
	// No further changes are accepted on the returned copy once sealed.
	Seal() Sealable

	// IsSealed checks if a sealable instance is sealed.
	IsSealed() bool
}

func assertNotSealed(s Sealable) {
	if s.IsSealed() {
		panic("type is sealed so it can't be modified")
	}
}
