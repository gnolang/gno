package book

//----------------------------------------
// Committee

// XXX incomplete.
type Committee struct {

	// full addr.
	faddr std.FullAddress

	// each account has an associated pab
	pab *PrivActionBook
}

// A Committee is an actor
var _ Actor = *Committee{}

func NewCommittee(pkgPath, subpath string) *Committee {
	return &Committee{
		faddr: xxx,
		pab:   xxx,
	}
}

func (cm *Committee) GetAddress() std.Address {
	XXX
}
