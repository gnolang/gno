package cla

import (
	"gno.land/p/moul/helplink"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/nt/ufmt/v0"
)

func Render(path string) string {
	out := md.H1("Contributor License Agreement (CLA)")

	out += md.Paragraph("A Contributor License Agreement (CLA) must be signed before deploying packages.")
	out += md.Paragraph(
		"The Agreement governs Contributions uploaded, published, or made available " +
			"for execution on the Gno.land blockchain network, and the " +
			"related software and repositories used to publish such Contributions.",
	)

	if requiredHash == "" {
		out += md.HorizontalRule()
		out += md.H2("Status")
		out += md.Paragraph(md.Bold("CLA enforcement is currently DISABLED."))
		out += md.Paragraph("All package deployments are allowed.")
		return out
	}

	out += md.HorizontalRule()
	out += md.H2("Status")
	out += md.Paragraph(md.Bold("CLA enforcement is ENABLED"))

	if claURL != "" {
		out += md.Paragraph("You can read the full agreement here: " + md.Link(claURL, claURL))
	}

	table := mdtable.Table{Headers: []string{"", ""}}
	table.Append([]string{md.Bold("Required Hash"), md.InlineCode(requiredHash)})
	table.Append([]string{md.Bold("Signers"), ufmt.Sprintf("%d contributor(s)", signatures.Size())})
	out += table.String()

	out += md.H3("Actions")
	out += md.Paragraph(helplink.Func("Sign CLA", "Sign", "hash", requiredHash))
	return out
}
