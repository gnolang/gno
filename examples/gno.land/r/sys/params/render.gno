package params

import (
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/nt/ufmt/v0"
)

// Render shows who, other than GovDAO, may currently write a chain parameter.
//
// This realm had no Render, so the only way to see a delegation was to know it
// existed and query for it by name. That is the wrong shape for state that
// grants a capability: someone auditing the chain should be able to see, in one
// place, whether any parameter is delegated and to whom.
//
// Values are rendered without escaping, which is safe here because none of them
// is free-form. A delegate path has passed assertDelegatePath, which permits
// only lowercase letters, digits and a few separators; the addresses come from
// the parameter itself, which the chain validates as bech32; and the valset
// realm is a compile-time constant.
func Render(path string) string {
	out := md.H1("Chain parameter delegation")
	out += md.Paragraph(
		"Chain parameters are written by GovDAO proposal. A parameter may also " +
			"be delegated to one other realm, which can then manage it without a " +
			"vote. GovDAO keeps full control of every parameter and can withdraw " +
			"a delegation at any time.",
	)

	out += md.HorizontalRule()
	out += md.H2("run_submitters")
	out += md.Paragraph("Addresses allowed to send MsgRun, which runs code the sender supplies.")

	mgr := RunSubmittersManager()
	table := mdtable.Table{Headers: []string{"", ""}}
	if mgr == "" {
		table.Append([]string{md.Bold("Delegated to"), "nobody, only GovDAO may change it"})
	} else {
		table.Append([]string{md.Bold("Delegated to"), md.InlineCode(mgr)})
	}

	listed := GetRunSubmitters()
	table.Append([]string{md.Bold("Addresses allowed"), ufmt.Sprintf("%d", len(listed))})
	out += table.String()

	if mgr != "" {
		// Which entries the delegate may remove again. Anything it did not add
		// is GovDAO's to remove, including whatever the chain started with.
		granted := 0
		for _, a := range listed {
			if RunSubmittersGrantedBy(address(a)) {
				granted++
			}
		}
		out += md.Paragraph(ufmt.Sprintf(
			"%d of those were added by the delegate, and are the only ones it may remove.",
			granted))
	}

	if len(listed) > 0 {
		items := make([]string, 0, len(listed))
		for _, a := range listed {
			items = append(items, md.InlineCode(a))
		}
		out += md.BulletList(items)
	}

	out += md.HorizontalRule()
	out += md.H2("node:valset")
	out += md.Paragraph(
		"Validator set parameters are written by one realm, fixed in this " +
			"realm's source. Unlike a delegation this cannot be repointed by a " +
			"vote, and GovDAO cannot write these keys itself.",
	)
	vtable := mdtable.Table{Headers: []string{"", ""}}
	vtable.Append([]string{md.Bold("Writable only by"), md.InlineCode(valsetAuthorizedRealm)})
	out += vtable.String()

	return out
}
