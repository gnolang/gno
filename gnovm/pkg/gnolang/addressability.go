package gnolang

type Addressability int

const (
	AddressabilityNotApplicable Addressability = iota
	AddressabilitySatisfied
	AddressabilityUnsatisfied
)
