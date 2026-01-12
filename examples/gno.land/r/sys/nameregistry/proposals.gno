package nameregistry

import (
	"strings"

	"gno.land/p/moul/md"

	"gno.land/r/gov/dao"
)

// ProposeConstraintsReset creates a new GovDAO proposal to reset name constraints.
// Reseting invalidates all existing check contrainst by replacing them with the new ones.
func ProposeConstraintsReset(checkers ...Checker) dao.ProposalRequest {
	if len(checkers) == 0 {
		panic("proposal requires at least one name checker")
	}

	var desc strings.Builder
	desc.WriteString(md.Paragraph("Proposal replaces current name registry realm constrains by:"))
	for _, c := range checkers {
		desc.WriteString(md.BulletItem(c.Description()))
	}

	cb := func(realm) error {
		constraints = checkers
		return nil
	}

	return dao.NewProposalRequest(
		"Reset name registry constraints",
		desc.String(),
		dao.NewSimpleExecutor(cb, ""),
	)
}

// ProposeConstraintsAppend creates a new GovDAO proposal to append new name constraints.
func ProposeConstraintsAppend(checkers ...Checker) dao.ProposalRequest {
	if len(checkers) == 0 {
		panic("proposal requires at least one name checker")
	}

	var desc strings.Builder
	desc.WriteString(md.Paragraph("Proposal appends new name registry realm constrains:"))
	for _, c := range checkers {
		desc.WriteString(md.BulletItem(c.Description()))
	}

	cb := func(realm) error {
		constraints = append(constraints, checkers...)
		return nil
	}

	return dao.NewProposalRequest(
		"Append new name registry constraints",
		desc.String(),
		dao.NewSimpleExecutor(cb, ""),
	)
}
