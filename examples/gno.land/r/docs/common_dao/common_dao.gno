package commondao

import (
	"gno.land/p/nt/commondao"
	"gno.land/p/moul/md"
)

const (
	DAO *commondao.CommonDAO
)

func CreateDao(title string) {
	caller := std.PreviousRealm().Address()

	// Creates DAO with Active and Finished proposal storage
	// Creates a member storage
	// Other options can be added, can be found in 'p/nt/commondao/commondao_options'
	DAO: commondao.New(
			commondao.WithName(title), // adds a name to the common DAO
			commondao.WithMember(caller), // Adds the creator of the DAO to its member storage
		),
}

func Render() string {
	out string

	out += md.H1("A CommonDao Example")

	return out
}