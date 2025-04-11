package pokemon

import (
	"errors"
	"std"
	"strconv"

	"gno.land/p/leo/utils"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
)

type GameState struct {
	myPokemon    *Pokemon
	otherPokemon *Pokemon
	otherTeam    [teamSize]Pokemon
}

func getGameState(params map[string]string, state string, me *UserData) (game GameState, err error) {
	if len(state)%2 == 1 {
		err = errors.New("state bad format")
		return
	}
	game.myPokemon = nil
	for i := range me.team {
		if me.team[i].name != "" {
			game.myPokemon = &me.team[i]
		}
	}
	if game.myPokemon == nil {
		err = errors.New("you need at least one Pokémon de play " + md.Bold(md.Link("Go to team selector", atm_path+":team?user="+params["user"])))
		return
	}
	for i := 0; i < teamSize; i += 1 {
		pokemonId := state[i] - '0'
		if int(pokemonId) >= len(pokemonList) {
			err = errors.New("invalid enemy : " + string(state[i]))
			return
		}
		game.otherTeam[i] = pokemonList[pokemonId].deepCopy()
	}
	game.otherPokemon = &game.otherTeam[0]
	for i := teamSize; i < len(state); i += 2 {
		actionParam := state[i+1] - '0'
		switch state[i] {
		case 's':
			if me.team[actionParam].name != "" {
				game.myPokemon = &me.team[actionParam]
			} else {
				err = errors.New("invalid swap : " + string(state[i+1]))
			}
		case 'a':
			if game.myPokemon != nil && game.myPokemon.pv > 0 && game.myPokemon.attacks[actionParam].pc > 0 {
				game.myPokemon.attacks[actionParam].pc -= 1
				game.otherPokemon.pv -= game.myPokemon.attacks[actionParam].damage
			} else {
				err = errors.New("invalid ally attack : " + string(state[i+1]))
			}
		case 'i':
			if game.myPokemon != nil && len(me.inventory) > int(actionParam) {
				itemData[me.inventory[actionParam]].applyer(game.myPokemon)
				me.inventory = append(me.inventory[:actionParam], me.inventory[actionParam+1:]...)
			} else {
				err = errors.New("invalid item use : " + string(state[i+1]))
			}
		case 'o':
			if game.otherPokemon != nil && game.otherPokemon.pv > 0 && game.otherPokemon.attacks[actionParam].pc > 0 {
				game.otherPokemon.attacks[actionParam].pc -= 1
				game.myPokemon.pv -= game.otherPokemon.attacks[actionParam].damage
			} else {
				err = errors.New("invalid enemy attack : " + string(state[i+1]))
			}
		default:
			err = errors.New("invalid action : " + string(state[i]))
		}
	}
	return
}

func getFightLink(params map[string]string) string {
	return atm_path + ":fight?user=" + params["user"] + "&state=" + params["state"]
}

func renderTeamSwap(game GameState, team *[teamSize]Pokemon, params map[string]string) (rend string) {
	have_pokemon := false
	table := mdtable.Table{Headers: []string{"Swap for "}, Rows: [][]string{[]string{"HP"}}}
	for i, p := range team {
		if p.name == "" {
			continue
		}
		have_pokemon = true
		if &team[i] == game.myPokemon {
			table.Headers = append(table.Headers, "Selected : "+p.name+md.Image(p.name, p.assetPath))
		} else {
			link := getFightLink(params) + "s" + strconv.Itoa(i)
			table.Headers = append(table.Headers, md.Link(p.name, link)+md.InlineImageWithLink(p.name, p.assetPath, link))
		}
		table.Rows[0] = append(table.Rows[0], strconv.Itoa(p.pv)+" / "+strconv.Itoa(p.getDefault().pv))
	}
	if !have_pokemon {
		rend += "Team : " + md.Bold("None !") + nl
	} else {
		rend += nl + table.String()
	}
	return
}

func renderInventoryUse(inventory []ItemType, params map[string]string) (rend string) {
	if len(inventory) != 0 {
		table := mdtable.Table{Headers: []string{"Item", "Description"}}
		for _, e := range inventory {
			table.Append([]string{md.Link("Use "+itemData[e].name, getFightLink(params)+"i"+strconv.Itoa(int(e))), itemData[e].description})
		}
		rend += nl + table.String()
	} else {
		rend += "Inventory : " + md.Bold("Empty !") + nl
	}
	return
}

func renderFightPokemon(p *Pokemon, params map[string]string, canAttack bool) (rend string) {
	if p == nil {
		rend = "Pokémon is NULL ??"
		return
	}
	rend += md.H3(p.name)

	rend += md.Image(p.name, p.assetPath)
	rend += md.Bold(strconv.Itoa(p.pv)+" / "+strconv.Itoa(p.getDefault().pv)) + " HP\n"

	table := mdtable.Table{Headers: []string{"ATTACK"},
		Rows: [][]string{
			[]string{md.Bold("Damage")},
			[]string{md.Bold("PC")},
		},
	}
	if !canAttack {
		table.Headers[0] = ""
	}
	for i, attack := range p.attacks {
		if canAttack {
			table.Headers = append(table.Headers, md.Link(attack.name, getFightLink(params)+"a"+strconv.Itoa(i)))
		} else {
			table.Headers = append(table.Headers, attack.name)
		}
		table.Rows[0] = append(table.Rows[0], strconv.Itoa(attack.damage))
		table.Rows[1] = append(table.Rows[1], strconv.Itoa(attack.pc)+" / "+strconv.Itoa(p.getDefault().attacks[i].pc))
	}
	rend += table.String()
	return
}

func renderPokemon1v1(game GameState, params map[string]string) (rend string) {
	rend += "<gno-columns>" + nl
	rend += renderFightPokemon(game.myPokemon, params, true)
	rend += nl + "|||" + nl
	rend += renderFightPokemon(game.otherPokemon, params, false)
	rend += nl + "</gno-columns>"
	return
}

func initEnemy(state string) string {
	if len(state) >= teamSize {
		return state
	}
	res := ""
	for i := 0; i < teamSize; i++ {
		res += strconv.Itoa(rand.Next() % len(pokemonList))
	}
	return res
}

func renderPokemonFight(params map[string]string) (rend string) {
	rend += md.H2("Fight !")
	idx := utils.AddressSliceIndex(pokemonDAO.GetMembers(), func(_ int, member std.Address) bool {
		return std.Address(params["user"]) == member
	})
	me := &pokemonDAO.userData[idx]
	params["state"] = initEnemy(params["state"])
	game, err := getGameState(params, params["state"], me)
	if err != nil {
		rend += md.Italic(err.Error()) + nl
		return
	}
	rend += renderTeamSwap(game, &me.team, params) + nl
	rend += renderInventoryUse(me.inventory, params) + nl
	rend += renderPokemon1v1(game, params) + nl
	rend += md.Link("Give up and go back to lobby", atm_path+":lobby?user="+params["user"])
	return
}
