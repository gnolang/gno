package pokemon

import (
	"gno.land/p/demo/mux"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
)

// pathDescription maps RenderPage to descriptions for navigation links
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

// renderWallet displays the player's Pokedollar balance
func renderWallet(money uint) string {
	return "Wallet : " + renderPokedollar(int(money)) + nl
}

// renderSelectedPokemon displays the player's selected Pokémon team in a table
func renderSelectedPokemon(team [teamSize]Pokemon) string {
	rend := "Selected Pokémons : "
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
		return rend + nl + table.String()
	}
	return rend + md.Bold("None !") + nl
}

// renderInventory displays the player's inventory with item details
func renderInventory(inventory []ItemType) string {
	if len(inventory) == 0 {
		return "Inventory : " + md.Bold("Empty !") + nl
	}
	table := mdtable.Table{Headers: []string{"Item", "Description", "Price"}}
	for _, e := range inventory {
		table.Append([]string{itemData[e].name, itemData[e].description, renderPokedollar(itemData[e].price)})
	}
	return "Inventory" + nl + table.String()
}

// renderMe displays the player's wallet, inventory, and selected Pokémon
func renderMe(user string) string {
	me, userExist := pokemonDAO.Get(user)
	if !userExist {
		return "Unkown user"
	}
	return renderWallet(me.(*UserData).pokedollar) +
		renderInventory(me.(*UserData).inventory) + nl +
		renderSelectedPokemon(me.(*UserData).team) + nl
}

// renderMap generates navigation links for the game pages
func renderMap(user string) (rend string) {
	for k, v := range pathDescription {
		link := atmPath + ":" + string(k) + "?user=" + user
		rend += " - " + md.Link(string(k), link)
		rend += " : " + v + nl
	}
	return
}

// renderPokemonLobby renders the main lobby page with navigation and player info
func renderPokemonLobby(rend *mux.ResponseWriter, req *mux.Request) {
	user := req.Query.Get("user")

	rend.Write(md.H2("Lobby"))

	rend.Write(md.H3("Where do you want to go ?"))
	rend.Write(renderMap(user))

	rend.Write(md.H3("Me"))
	rend.Write(renderMe(user))
	return
}
