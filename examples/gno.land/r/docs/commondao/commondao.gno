package commondao

import (
	"std"

	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/nt/commondao"
)

const (
	chaegnal = "g1dzzx5qwcmawqnhhd76ehkteas5v2dg42hqnd7g"
	leon     = "g125em6arxsnj49vx35f0n0z34putv5ty3376fg5"
	manfred  = "g1manfred47kzduec920z88wfr64ylksmdcedlf5"
)

// DAO is the instance of the CommonDAO used by this realm
var DAO *commondao.CommonDAO

// CreateDao initializes a new DAO instance
func CreateDao(title string) {
	// Creates a DAO with Active and Finished proposal storage
	// Also, it creates a member storage
	// Other options can be added, all of which can be found in 'p/nt/commondao/commondao_options'
	DAO = commondao.New(
		commondao.WithName(title),      // adds a name to the common DAO
		commondao.WithMember(chaegnal), // adds the creator of the DAO to its member storage
		commondao.WithDescription("An example DAO"),
	)
}

func Render(_ string) string {
	out := md.H1("A CommonDao Example")

	out += md.Paragraph("This realm is meant to demonstrate how to use the `[/p/nt/commondao](/p/nt/commondao)` pure package.")
	out += md.Paragraph("The commondao package is meant to act as a DAO factory to use in any realms easily.")

	out += md.Paragraph("In the source code, you'll find out how to create a basic DAO with where to look for more informations about more advanced options")

	initDao()

	out += md.H2("DAO Overview")
	out += md.Bold("Title:") + " " + DAO.Name() + "\n\n"
	out += md.Bold("Description:") + " " + DAO.Description() + "\n\n"

	out += md.H3("DAO's members")
	out += displayMembers()
	return out
}

// initDao creates the example DAO of this realm
func initDao() {
	// Creation of the DAO with the given name
	CreateDao("CommonDAO Documentation Realm")

	// Adds two new members to the DAO
	DAO.Members().Add(leon)
	DAO.Members().Add(manfred)
}

func displayMembers() string {
	table := mdtable.Table{
		Headers: []string{"Address"},
	}

	members := DAO.Members()
	members.IterateByOffset(0, members.Size(), func(adr std.Address) bool {
		table.Append([]string{adr.String()})
		return false
	})

	return table.String()
}
