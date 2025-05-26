package pokemon

import (
	"std"

	"gno.land/p/demo/mux"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
)

var (
	pathDescription = map[RenderPage]string{
		pokemonLogin:        "To swap account, or delete your actual one",
		pokemonLobby:        "This page",
		pokemonTeamSelector: "To select a team of 6 Pokémon",
		pokemonFight:        "Fight against other Pokémon trainer",
		pokemonShop:         "A shop to buy items with your Pokedollars",
		pokemonRank:         "Hall of Fame for the better Pokémon trainers !",
	}
)

func renderWallet(money uint) string {
	return "Wallet : " + renderPokedollar(int(money)) + nl
}

func renderSelectedPokemon(team [teamSize]Pokemon) (rend string) {
	rend += "Selected Pokémons : "
	have_pokemon := false
	table := mdtable.Table{Headers: []string{}, Rows: [][]string{[]string{}}}
	for _, p := range team {
		if p.name == "" {
			continue
		}
		have_pokemon = true
		table.Headers = append(table.Headers, p.name)
		table.Rows[0] = append(table.Rows[0], md.Image(p.name, p.assetPath))
	}
	if have_pokemon {
		rend += nl + table.String()
	} else {
		rend += md.Bold("None !") + nl
	}
	return
}

func renderInventory(inventory []ItemType) (rend string) {
	rend += "Inventory : "
	if len(inventory) != 0 {
		table := mdtable.Table{Headers: []string{"Item", "Description", "Price"}}
		for _, e := range inventory {
			table.Append([]string{itemData[e].name, itemData[e].description, renderPokedollar(itemData[e].price)})
		}
		rend += nl + table.String()
	} else {
		rend += md.Bold("Empty !") + nl
	}
	return
}

func renderMe(user string) (rend string) {
	idx := addressSliceIndex(pokemonDAO.GetMembers(), func(_ int, member std.Address) bool {
		return std.Address(user) == member
	})
	me := pokemonDAO.userData[idx]
	rend += renderWallet(me.pokedollar)
	rend += renderInventory(me.inventory) + nl
	rend += renderSelectedPokemon(me.team) + nl
	return
}

func renderMap(user string) (rend string) {
	for k, v := range pathDescription {
		link := atmPath + ":" + string(k) + "?user=" + user
		rend += " - " + md.Link(string(k), link)
		rend += " : " + v + nl
	}
	return
}

func renderPokemonLobby(rend *mux.ResponseWriter, req *mux.Request) {
	user := req.Query.Get("user")

	rend.Write(md.H2("Lobby"))

	rend.Write(md.H3("Where do you want to go ?"))
	rend.Write(renderMap(user))

	rend.Write(md.H3("Me"))
	rend.Write(renderMe(user))
	return
}
