package commondao

import (
	"errors"

	"gno.land/p/nt/commondao"
)

var ErrInvalidProposal = errors.New("invalid proposal definition")

func assertIsValidProposalDefinition(p commondao.Validable) {
	if err := p.Validate(); err != nil {
		panic(err)
	}
}
