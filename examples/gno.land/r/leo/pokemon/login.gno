package pokemon

import (
	"std"

	"gno.land/p/demo/mux"
	"gno.land/p/leo/dao"
	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"
	"gno.land/r/sys/users"
)

type UserData struct {
	team       [teamSize]Pokemon
	inventory  []ItemType
	pokedollar uint
}

type PokemonDAO struct {
	dao.DAO
	userData []UserData
}

var (
	pokemonDAO = PokemonDAO{}
)

// defaultUserData creates a new user with an empty team and 500 Pokedollars
func defaultUserData() UserData {
	return UserData{team: [teamSize]Pokemon{}, inventory: []ItemType{}, pokedollar: 500}
}

// Login registers a new user in the Pokémon game
func Login() {
	crossing()
	usr := std.PreviousRealm().Address()
	pokemonDAO.AsserUserIsNotMember(usr)
	pokemonDAO.userData = append(pokemonDAO.userData, defaultUserData())
	pokemonDAO.AddMember(usr)
}

// DeleteMyAccount removes the user's account and data
func DeleteMyAccount() {
	crossing()
	usr := std.PreviousRealm().Address()
	pokemonDAO.AsserUserIsMember(usr)
	idx := addressSliceIndex(pokemonDAO.GetMembers(), func(_ int, member std.Address) bool {
		return usr == member
	})
	pokemonDAO.userData = append(pokemonDAO.userData[:idx], pokemonDAO.userData[idx+1:]...)
	pokemonDAO.DeleteMember(usr)
}

// renderPokemonLogin renders the login page with options to create or delete an account
func renderPokemonLogin(rend *mux.ResponseWriter, _ *mux.Request) {
	members := pokemonDAO.GetMembers()
	rend.Write(md.H2("Are you a already Pokémon Trainer ?"))
	rend.Write(md.Link("Get me a brand new Pokémon Licence !", txlink.Call("Login")) + nl)
	rend.Write(md.Link("Delete my existing Pokémon Licence", txlink.Call("DeleteMyAccount")) + nl)
	rend.Write(md.H2("Select your account"))
	for i, e := range pokemonDAO.userData {
		name := string(members[i])
		data := users.ResolveAddress(members[i])
		if data != nil {
			name = data.Name()
		}
		link := atmPath + ":" + string(pokemonLobby) + "?user=" + string(members[i])
		rend.Write(" - " + md.Link(name, link))
		rend.Write(" : " + renderPokedollar(int(e.pokedollar)) + nl)
	}
	return
}
