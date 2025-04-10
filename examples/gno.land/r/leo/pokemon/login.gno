package pokemon

import (
	"std"
	"strconv"

	"gno.land/p/leo/dao"
	"gno.land/p/leo/renderer"
	"gno.land/p/leo/utils"
	"gno.land/p/moul/txlink"
)

type ItemType int

const (
	SmallPotion ItemType = iota
	BigPotion
	MaxPotion
	Revive
	MaxRevive
	Elixer
	MaxElixer
)

type Item struct {
	price       int
	name        string
	description string
}

type UserData struct {
	team       [6]Pokemon
	inventory  []ItemType
	pokedollar uint
}

type PokemonDAO struct {
	dao.DAO
	userData []UserData
}

var (
	pokemonDAO = PokemonDAO{}
	itemData   = map[ItemType]Item{
		SmallPotion: {300, "Small Potion", "Restores 20 HP for a Pokémon"},
		BigPotion:   {700, "Big Potion", "Restores 60 HP for a Pokémon"},
		MaxPotion:   {2500, "Max Potion", "Fully recovers the HP of a Pokémon"},
		Revive:      {1500, "Revive", "Revive a Pokémon and restores half its HP"},
		MaxRevive:   {3500, "Max Revive", "Revive a Pokémon and fully recovers its HP"},
		Elixer:      {1000, "Elixer", "Restores 10 PP of each ability of one Pokémon"},
		MaxElixer:   {2000, "Max Elixer", "Restores all the PP of each ability of one Pokémon"},
	}
)

func defaultUserData() UserData {
	return UserData{team: [6]Pokemon{}, inventory: []ItemType{}, pokedollar: 500}
}

func Login() {
	usr := std.PreviousRealm().Address()
	pokemonDAO.AsserUserIsNotMember(usr)

	pokemonDAO.userData = append(pokemonDAO.userData, defaultUserData())
	pokemonDAO.AddMember(usr)
}

func DeleteMyAccount() {
	usr := std.PreviousRealm().Address()
	pokemonDAO.AsserUserIsMember(usr)

	idx := utils.AddressSliceIndex(pokemonDAO.GetMembers(), func(_ int, member std.Address) bool {
		return usr == member
	})
	pokemonDAO.userData = append(pokemonDAO.userData[:idx], pokemonDAO.userData[idx+1:]...)
	pokemonDAO.DeleteMember(usr)
}

func renderPokemonLogin(params map[string]string) (rend string) {
	rend += renderer.Header(2, "Are you a already Pokémon Trainer ?")
	rend += renderer.Link("Get me a brand new Pokémon Licence !", txlink.Call("Login")) + nl
	rend += renderer.Link("Delete my existing Pokémon Licence", txlink.Call("DeleteMyAccount")) + nl
	rend += renderer.Header(2, "Select your account")
	for i, e := range pokemonDAO.userData {
		link := ":" + string(pokemonLobby) + "?user=" + string(pokemonDAO.GetMembers()[i])
		rend += " - " + renderer.Link(string(pokemonDAO.GetMembers()[i]), link)
		rend += ": " + strconv.FormatUint(uint64(e.pokedollar), 10) + " " + nl
	}
	return
}
