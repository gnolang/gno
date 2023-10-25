package book

//----------------------------------------
// Committee

// XXX incomplete.
type Committee struct {
	// each account has an associated pab
	// XXX or, accounts have devices, each with a pab.
	pab *PrivActionBook
}

// A Committee is an actor
var _ Actor = *Committee{}
