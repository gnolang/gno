package nameregistry

import (
	"strings"

	"gno.land/p/moul/md"
	"gno.land/p/nt/ufmt"

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

// ProposeNamesSet creates a new GovDAO proposal to set reserved, blacklisted or custom names.
func ProposeNamesSet(nameType string, names ...string) dao.ProposalRequest {
	if len(names) == 0 {
		panic("list of names must not be empty")
	}

	if nameType != TypeReserved && nameType != TypeBlacklisted && nameType != TypeCustom {
		panic("invalid name type: " + nameType)
	}

	cb := func(realm) error {
		SetNames(cross, nameType, names...)
		return nil
	}

	return dao.NewProposalRequest(
		ufmt.Sprintf("Initialize %s names", nameType),
		ufmt.Sprintf("Proposal initializes %d %s names", len(names), nameType),
		dao.NewSimpleExecutor(cb, ""),
	)
}

// ProposeNamesAdd creates a new GovDAO proposal to add reserved, blacklisted or custom names.
func ProposeNamesAdd(nameType string, names ...string) dao.ProposalRequest {
	if len(names) == 0 {
		panic("list of names must not be empty")
	}

	if nameType != TypeReserved && nameType != TypeBlacklisted && nameType != TypeCustom {
		panic("invalid name type: " + nameType)
	}

	cb := func(realm) error {
		AddNames(cross, nameType, names...)
		return nil
	}

	return dao.NewProposalRequest(
		ufmt.Sprintf("Add new %s names", nameType),
		ufmt.Sprintf("Proposal adds %d %s new names", len(names), nameType),
		dao.NewSimpleExecutor(cb, ""),
	)
}
