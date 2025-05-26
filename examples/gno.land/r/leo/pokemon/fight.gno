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

const (
	gameContinue int = iota
	gameWin
	gameLoose
)

type GameState struct {
	myPokemon    *Pokemon
	otherPokemon *Pokemon
	otherTeam    [teamSize]Pokemon
	state        int
}

func checkGameEnd(myTeam [teamSize]Pokemon, game *GameState) bool {
	game.state = gameWin
	for j := 0; j < teamSize; j++ {
		if game.otherTeam[j].pv > 0 {
			game.state = gameContinue
			break
		}
	}
	if game.state == gameWin {
		return true
	}
	game.state = gameLoose
	for j := 0; j < teamSize; j++ {
		if myTeam[j].pv > 0 {
			game.state = gameContinue
			break
		}
	}
	if game.state == gameLoose {
		return true
	}
	return false
}
func getGameState(state string, me *UserData) (game GameState, err error) {
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
		err = errors.New("you need at least one Pokémon to play")
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
	if checkGameEnd(me.team, &game) {
		return
	}
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
				if game.otherPokemon.pv <= 0 {
					for i := 0; i < teamSize; i++ {
						if game.otherTeam[i].pv > 0 {
							game.otherPokemon = &game.otherTeam[i]
							break
						}
					}
				}
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
		case 'e':
			if game.otherPokemon != nil && game.otherPokemon.pv > 0 && game.otherPokemon.attacks[actionParam].pc > 0 {
				game.otherPokemon.attacks[actionParam].pc -= 1
				game.myPokemon.pv -= game.otherPokemon.attacks[actionParam].damage
			} else {
				err = errors.New("invalid enemy attack : " + string(state[i+1]))
			}
		default:
			err = errors.New("invalid action : " + string(state[i]))
		}
		if checkGameEnd(me.team, &game) {
			return
		}
	}
	return
}

func getFightLink(user, state string) string {
	return atmPath + ":fight?user=" + user + "&state=" + state
}

func getEnnemyNextAttack(p *Pokemon) string {
	can_attack := false
	for _, a := range p.attacks {
		if a.pc > 0 {
			can_attack = true
		}
	}
	if !can_attack {
		return ""
	}
	for {
		attack_id := r.Int() % len(p.attacks)
		if p.attacks[attack_id].pc > 0 {
			return "e" + strconv.Itoa(attack_id)
		}
	}
	return ""
}

func renderTeamSwap(game GameState, team *[teamSize]Pokemon, user, state string) (rend string) {
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
			link := getFightLink(user, state) + "s" + strconv.Itoa(i)
			table.Headers = append(table.Headers, md.Link(p.name, link)+md.InlineImageWithLink(p.name, p.assetPath, link))
		}
		if p.pv > 0 {
			table.Rows[0] = append(table.Rows[0], strconv.Itoa(p.pv)+" / "+strconv.Itoa(p.getDefault().pv))
		} else {
			table.Rows[0] = append(table.Rows[0], strconv.Itoa(0)+" / "+strconv.Itoa(p.getDefault().pv))
		}
	}
	if !have_pokemon {
		rend += "Team : " + md.Bold("None !") + nl
	} else {
		rend += nl + table.String()
	}
	return
}

func renderInventoryUse(inventory []ItemType, user, state string) (rend string) {
	if len(inventory) != 0 {
		table := mdtable.Table{Headers: []string{"Item", "Description"}}
		for i, e := range inventory {
			table.Append([]string{md.Link("Use "+itemData[e].name, getFightLink(user, state)+"i"+strconv.Itoa(i)), itemData[e].description})
		}
		rend += nl + table.String()
	} else {
		rend += "Inventory : " + md.Bold("Empty !") + nl
	}
	return
}

func renderFightPokemon(p *Pokemon, user, state string, canAttack bool, otherP *Pokemon) (rend string) {
	if p == nil {
		rend = "Pokémon is NULL ??"
		return
	}
	rend += md.H3(p.name)

	rend += md.Image(p.name, p.assetPath)
	if p.pv > 0 {
		rend += md.Bold(strconv.Itoa(p.pv)+" / "+strconv.Itoa(p.getDefault().pv)) + " HP\n"
	} else {
		rend += md.Bold(strconv.Itoa(0)+" / "+strconv.Itoa(p.getDefault().pv)) + " HP\n"
	}

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
		if canAttack && attack.pc > 0 && p.pv > 0 {
			table.Headers = append(table.Headers, md.Link(attack.name, getFightLink(user, state)+"a"+strconv.Itoa(i)+getEnnemyNextAttack(otherP)))
		} else {
			table.Headers = append(table.Headers, attack.name)
		}
		table.Rows[0] = append(table.Rows[0], strconv.Itoa(attack.damage))
		table.Rows[1] = append(table.Rows[1], strconv.Itoa(attack.pc)+" / "+strconv.Itoa(p.getDefault().attacks[i].pc))
	}
	rend += table.String()
	return
}

func renderPokemon1v1(game GameState, user, state string) (rend string) {
	rend += "<gno-columns>" + nl
	rend += renderFightPokemon(game.myPokemon, user, state, true, game.myPokemon)
	rend += nl + "|||" + nl
	rend += renderFightPokemon(game.otherPokemon, user, state, false, nil)
	rend += nl + "</gno-columns>"
	return
}

func initEnemy(state string) string {
	if len(state) >= teamSize {
		return state
	}
	res := ""
	for i := 0; i < teamSize; i++ {
		res += strconv.Itoa(r.Int() % len(pokemonList))
	}
	return res
}

func ApplyGameEnd(state string) {
	usr := std.PreviousRealm().Address()
	pokemonDAO.AsserUserIsMember(usr)
	idx := addressSliceIndex(pokemonDAO.GetMembers(), func(_ int, member std.Address) bool {
		return usr == member
	})
	me := &pokemonDAO.userData[idx]

	game, err := getGameState(state, me)
	if err != nil {
		panic("bad state : " + err.Error())
	}
	switch game.state {
	case gameContinue:
		panic("game is not finished yet")
	case gameLoose:
		if me.pokedollar >= 200 {
			me.pokedollar -= 200
		} else {
			me.pokedollar = 0
		}
	case gameWin:
		me.pokedollar += 1000
	}
	for i, e := range me.team {
		if e.name == "" {
			continue
		}
		me.team[i] = e.getDefault().deepCopy()
	}
}

func renderEndGame(game GameState, me *UserData, user string, state string) (rend string) {
	if game.state == gameWin {
		rend += md.H3("You WIN ! GGWP")
		rend += md.Link("Get your reward !", txlink.Call("ApplyGameEnd", "state", state)) + nl
		rend += md.Link("Go back to lobby", atmPath+":lobby?user="+user) + nl
	}
	if game.state == gameLoose {
		rend += md.H3("You LOOSE...")
		rend += md.Link("Accept !", txlink.Call("ApplyGameEnd", "state", state)) + nl
		rend += md.Link("Go back to lobby", atmPath+":lobby?user="+user) + nl
	}
	return
}

func renderPokemonFight(rend *mux.ResponseWriter, req *mux.Request) {
	rend.Write(md.H2("Fight !"))
	user := req.Query.Get("user")
	idx := addressSliceIndex(pokemonDAO.GetMembers(), func(_ int, member std.Address) bool {
		return std.Address(user) == member
	})
	me := &pokemonDAO.userData[idx]
	req.Query.Set("state", initEnemy(req.Query.Get("state")))
	state := req.Query.Get("state")
	game, err := getGameState(state, me)
	if err != nil {
		rend.Write(md.Italic(err.Error()) + nl)
		rend.Write(md.Link("Go back to lobby", atmPath+":lobby?user="+user))
		return
	}
	if game.state != gameContinue {
		rend.Write(renderEndGame(game, me, user, state) + nl)
		return
	}
	rend.Write(renderTeamSwap(game, &me.team, user, state) + nl)
	rend.Write(renderInventoryUse(me.inventory, user, state) + nl)
	rend.Write(renderPokemon1v1(game, user, state) + nl)
	rend.Write(md.Link("Give up and go back to lobby", atmPath+":lobby?user="+user))
	return
}
