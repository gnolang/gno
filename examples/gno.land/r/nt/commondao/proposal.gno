package commondao

import (
	"errors"

	"gno.land/p/nt/commondao"
)

// i'd expect functions related to (default) proposals i can
// on the DAO object, or something related to proposals in general

var ErrInvalidProposal = errors.New("invalid proposal definition")

func assertIsValidProposalDefinition(p commondao.Validable) {
	if err := p.Validate(); err != nil {
		panic(err)
	}
}
