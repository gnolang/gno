package pokemon

import (
	"errors"
	"std"
	"strconv"

	"gno.land/p/demo/mux"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/txlink"
)

// SwapTeam updates the user's Pokémon team based on the provided team string
func SwapTeam(team string) {
	crossing()

	usr := string(std.PreviousRealm().Address())
	if !pokemonDAO.Has(usr) {
		panic("user is not member")
	}
	me, _ := pokemonDAO.Get(usr)
	if len(team) != teamSize {
		panic("a team contain 6 pokémon or empty slots")
	}
	for i := 0; i < len(team); i++ {
		if team[i] == 'p' {
			me.(*UserData).team[i] = Pokemon{}
		} else if team[i] >= '0' && team[i] <= '9' {
			me.(*UserData).team[i] = pokemonList[team[i]-'0'].deepCopy()
		} else {
			panic("team bad format !" + string(team[i]))
		}
	}
}

// getTeam validates the team string format for Pokémon selection
func getTeam(state string) (err error) {
	if len(state) != teamSize {
		err = errors.New("")
		return
	}
	for i := 0; i < len(state); i++ {
		if state[i] != 'p' && !(state[i] >= '0' && state[i] <= '9') {
			err = errors.New("invalid Pokémon : " + string(state[i]))
			return
		}
	}
	err = nil
	return
}

// getTeamLink generates a URL for the team selection page with user and team parameters
func getTeamLink(user, team string) string {
	return atmPath + ":team?user=" + user + "&team=" + team
}

// renderTeam renders the current team with swap options for the team selector page
func renderTeam(team string, user string, swap int) (rend string) {
	table := mdtable.Table{Headers: []string{}, Rows: [][]string{[]string{}, []string{}}}
	for i, p := range team {
		if p == 'p' {
			table.Headers = append(table.Headers, "Empty Slot")
			table.Rows[0] = append(table.Rows[0], " - ")
		} else {
			table.Headers = append(table.Headers, pokemonList[p-'0'].name)
			table.Rows[0] = append(table.Rows[0], md.Image(pokemonList[p-'0'].name, pokemonList[p-'0'].assetPath))
		}
		if i == swap {
			table.Rows[1] = append(table.Rows[1], "Swapping...")
		} else {
			table.Rows[1] = append(table.Rows[1], md.Link("Swap", getTeamLink(user, team)+"&swap="+strconv.Itoa(i)))
		}
	}
	rend += table.String()
	return
}

// renderPokemonList renders a Pokémon's details with a link to select it for the team
func renderPokemonList(p Pokemon, link string) (rend string) {
	rend += md.H3(p.name)

	rend += md.InlineImageWithLink(p.name, p.assetPath, link) + nl
	rend += md.Bold(strconv.Itoa(p.pv)) + " HP\n"

	table := mdtable.Table{Headers: []string{"ATTACK"},
		Rows: [][]string{
			[]string{md.Bold("Damage")},
			[]string{md.Bold("PC")},
		},
	}
	for _, attack := range p.attacks {
		table.Headers = append(table.Headers, attack.name)
		table.Rows[0] = append(table.Rows[0], strconv.Itoa(attack.damage))
		table.Rows[1] = append(table.Rows[1], strconv.Itoa(attack.pc))
	}
	rend += table.String()
	return
}

// renderPokemonTeam renders the team selection page with current team and available Pokémon
func renderPokemonTeam(rend *mux.ResponseWriter, req *mux.Request) {
	user := req.Query.Get("user")
	team := req.Query.Get("team")
	swapParam := req.Query.Get("swap")

	rend.Write(md.H2("Team selector"))
	me, user_exist := pokemonDAO.Get(user)
	if !user_exist {
		rend.Write(md.H3("Unkown user"))
		return
	}
	err := getTeam(team)
	if err != nil {
		if len(err.Error()) != 0 {
			rend.Write(md.Italic(err.Error()) + nl)
		}
		team = ""
		for _, p := range me.(*UserData).team {
			pokeId := p.getId()
			if pokeId == -1 {
				team += "p"
			} else {
				team += string(pokeId + '0')
			}
		}
	}
	swap, err := strconv.Atoi(swapParam)
	if err == nil && swap >= 0 && swap <= teamSize {
		rend.Write(renderTeam(team, user, swap) + nl)
		for i, p := range pokemonList {
			b := []byte(team)
			b[swap] = byte(i + '0')
			rend.Write(renderPokemonList(p, getTeamLink(user, string(b))))
		}
	} else {
		rend.Write(md.H3("Wich Pokémon do you want to swap ? (first pokemon is also your first during fight)"))
		rend.Write(renderTeam(team, user, -1) + nl)
		rend.Write(md.Link("Validate swap", txlink.Call("SwapTeam", "team", team)) + nl)
	}
	rend.Write(nl + md.Link("Back to lobby", atmPath+":lobby?user="+user))
	return
}
