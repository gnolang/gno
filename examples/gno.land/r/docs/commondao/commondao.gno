package commondao

import (
	"std"

	"gno.land/p/moul/md"
	"gno.land/p/nt/commondao"
)

// DAO is the instance of the CommonDAO used by this realm
var DAO *commondao.CommonDAO

// CreateDao initializes a new DAO instance
func CreateDao(title string) {
	caller := std.PreviousRealm().Address()

	// Creates DAO with Active and Finished proposal storage
	// Creates a member storage
	// Other options can be added, can be found in 'p/nt/commondao/commondao_options'
	DAO = commondao.New(
		commondao.WithName(title),    // adds a name to the common DAO
		commondao.WithMember(caller), // adds the creator of the DAO to its member storage
	)
}

func Render(_ string) string {
	out := md.H1("A CommonDao Example")

	out += md.Paragraph("This realm is meant to demonstrate how to use the [p/nt/commondao](/p/nt/commondao) pure package.")
	out += md.Paragraph("This package is meant to act as a DAO factory to use in any realms easily.")

	out += md.Paragraph("In the source code, you'll find out how to create a basic DAO with where to look for more informations about more advanced options")

	return out
}
