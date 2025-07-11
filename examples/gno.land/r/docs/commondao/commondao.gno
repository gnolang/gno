package commondao

import (
	"std"

	"gno.land/p/moul/md"
	"gno.land/p/nt/commondao"
)

var (
	DAO *commondao.CommonDAO
)

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

	return out
}
