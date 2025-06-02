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
	gameContinue int = iota // Game is ongoing
	gameWin                 // Player has won the game
	gameLoose               // Player has lost the game
)

var (
	stateActions = map[byte](func(game *GameState, me *UserData, actionParam byte) error){
		's': switchAction,      // switch pokemon
		'a': attackAction,      // use pc to attack enemy
		'i': itemUseAction,     // use an item from inventory
		'e': enemyAttackAction, // enemy attack our pokemon
	}
)

func switchAction(game *GameState, me *UserData, actionParam byte) error {
	if me.team[actionParam].name == "" {
		return errors.New("invalid swap")
	}
	game.myPokemon = &me.team[actionParam]
	return nil
}

func attackAction(game *GameState, me *UserData, actionParam byte) error {

	if game.myPokemon == nil || game.myPokemon.pv <= 0 || game.myPokemon.attacks[actionParam].pc <= 0 {
		return errors.New("invalid ally attack")
	}
	game.myPokemon.attacks[actionParam].pc -= 1
	game.otherPokemon.pv -= game.myPokemon.attacks[actionParam].damage
	if game.otherPokemon.pv <= 0 {
		game.otherPokemon.pv = 0
		for i := 0; i < teamSize; i++ {
			if game.otherTeam[i].pv > 0 {
				game.otherPokemon = &game.otherTeam[i]
				break
			}
		}
	}
	return nil
}

func itemUseAction(game *GameState, me *UserData, actionParam byte) error {
	if game.myPokemon == nil || len(me.inventory) <= int(actionParam) {
		return errors.New("invalid item use")
	}
	itemData[me.inventory[actionParam]].applyer(game.myPokemon)
	me.inventory = append(me.inventory[:actionParam], me.inventory[actionParam+1:]...)
	return nil
}

func enemyAttackAction(game *GameState, me *UserData, actionParam byte) error {
	if game.otherPokemon == nil || game.otherPokemon.pv <= 0 || game.otherPokemon.attacks[actionParam].pc <= 0 {
		return errors.New("invalid enemy attack")
	}
	game.otherPokemon.attacks[actionParam].pc -= 1
	game.myPokemon.pv -= game.otherPokemon.attacks[actionParam].damage
	if game.myPokemon.pv < 0 {
		game.myPokemon.pv = 0
	}
	return nil
}

// GameState holds the current state of a Pokémon battle
type GameState struct {
	myPokemon    *Pokemon
	otherPokemon *Pokemon
	otherTeam    [teamSize]Pokemon
	state        int
}

// checkGameEnd determines if the game has ended and updates the game state
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

// getGameState parses the game state string and initializes the battle state
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
		action, action_exist := stateActions[state[i]]
		if !action_exist {
			err = errors.New("invalid action : " + string(state[i]))
		} else {
			err = action(&game, me, actionParam)
		}
		if err != nil {
			return
		}
		if checkGameEnd(me.team, &game) {
			return
		}
	}
	return
}

// getFightLink generates a URL for the fight page with user and state parameters
func getFightLink(user, state string) string {
	return atmPath + ":fight?user=" + user + "&state=" + state
}

// getEnnemyNextAttack selects a random valid attack for the enemy Pokémon
func getEnnemyNextAttack(p *Pokemon) string {
	canAttack := false
	for _, a := range p.attacks {
		if a.pc > 0 {
			canAttack = true
		}
	}
	if !canAttack {
		return ""
	}
	for {
		attack_id := r.Int() % len(p.attacks)
		if p.attacks[attack_id].pc > 0 {
			return "e" + strconv.Itoa(attack_id)
		}
	}
}

// renderTeamSwap renders the team swap interface for the player
func renderTeamSwap(game GameState, team [teamSize]Pokemon, user, state string) string {
	table := mdtable.Table{Headers: []string{"Swap for "}, Rows: [][]string{[]string{"HP"}}}
	for i, poke := range team {
		if poke.name == "" {
			continue
		}
		headerName := ""
		if &team[i] == game.myPokemon {
			headerName = "Selected : " + poke.name + md.Image(poke.name, poke.assetPath)
		} else {
			link := getFightLink(user, state) + "s" + strconv.Itoa(i)
			headerName = md.Link(poke.name, link) + md.InlineImageWithLink(poke.name, poke.assetPath, link)
		}
		table.Headers = append(table.Headers, headerName)
		table.Rows[0] = append(table.Rows[0], strconv.Itoa(poke.pv)+" / "+strconv.Itoa(poke.getDefault().pv))
	}
	return nl + table.String()
}

// renderInventoryUse renders the player's inventory for item usage
func renderInventoryUse(inventory []ItemType, user, state string) string {
	if len(inventory) == 0 {
		return "Inventory : " + md.Bold("Empty !") + nl
	}
	table := mdtable.Table{Headers: []string{"Item", "Description"}}
	for i, e := range inventory {
		table.Append([]string{md.Link("Use "+itemData[e].name, getFightLink(user, state)+"i"+strconv.Itoa(i)), itemData[e].description})
	}
	return nl + table.String()
}

// renderFightPokemon renders a Pokémon's details and attack options for the fight page
func renderFightPokemon(p *Pokemon, user, state string, canAttack bool, otherP *Pokemon) string {
	if p == nil {
		return "Pokémon is NULL"
	}
	rend := md.H3(p.name)
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
		if canAttack && attack.pc > 0 && p.pv > 0 {
			table.Headers = append(table.Headers, md.Link(attack.name, getFightLink(user, state)+"a"+strconv.Itoa(i)+getEnnemyNextAttack(otherP)))
		} else {
			table.Headers = append(table.Headers, attack.name)
		}
		table.Rows[0] = append(table.Rows[0], strconv.Itoa(attack.damage))
		table.Rows[1] = append(table.Rows[1], strconv.Itoa(attack.pc)+" / "+strconv.Itoa(p.getDefault().attacks[i].pc))
	}
	return rend + table.String()
}

// renderPokemon1v1 renders a 1v1 battle view with both Pokémon
func renderPokemon1v1(game GameState, user, state string) (rend string) {
	rend += "<gno-columns>" + nl
	rend += renderFightPokemon(game.myPokemon, user, state, true, game.myPokemon)
	rend += nl + "|||" + nl
	rend += renderFightPokemon(game.otherPokemon, user, state, false, nil)
	rend += nl + "</gno-columns>"
	return
}

// initEnemy initializes the enemy team with random Pokémon if not already set
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

// ApplyGameEnd updates user data based on the game outcome
func ApplyGameEnd(state string) {
	crossing()
	usr := string(std.PreviousRealm().Address())
	if !pokemonDAO.Has(usr) {
		panic("user is not a member")
	}
	me, _ := pokemonDAO.Get(usr)
	game, err := getGameState(state, me.(*UserData))
	if err != nil {
		panic("bad state : " + err.Error())
	}
	switch game.state {
	case gameContinue:
		panic("game is not finished yet")
	case gameLoose:
		if me.(*UserData).pokedollar >= 200 {
			me.(*UserData).pokedollar -= 200
		} else {
			me.(*UserData).pokedollar = 0
		}
	case gameWin:
		me.(*UserData).pokedollar += 1000
	}
	for i, e := range me.(*UserData).team {
		if e.name == "" {
			continue
		}
		me.(*UserData).team[i] = e.getDefault().deepCopy()
	}
}

// renderEndGame renders the game outcome screen with reward or loss options
func renderEndGame(game GameState, me *UserData, user string, state string) (rend string) {
	if game.state == gameWin {
		rend += md.H3("You WIN ! GGWP")
		rend += md.Link("Get your reward !", txlink.Call("ApplyGameEnd", "state", state)) + nl
		rend += md.Link("Go back to lobby", atmPath+":lobby?user="+user) + nl
	} else {
		rend += md.H3("You LOSE...")
		rend += md.Link("Accept !", txlink.Call("ApplyGameEnd", "state", state)) + nl
		rend += md.Link("Go back to lobby", atmPath+":lobby?user="+user) + nl
	}
	return
}

// renderPokemonFight renders the main fight page with team, inventory, and battle view
func renderPokemonFight(rend *mux.ResponseWriter, req *mux.Request) {
	rend.Write(md.H2("Fight !"))
	user := req.Query.Get("user")
	me, userExist := pokemonDAO.Get(user)
	if !userExist {
		rend.Write(md.H3("Unkown user"))
		return
	}
	req.Query.Set("state", initEnemy(req.Query.Get("state")))
	state := req.Query.Get("state")
	game, err := getGameState(state, me.(*UserData))
	if err != nil {
		rend.Write(md.Italic(err.Error()) + nl)
		rend.Write(md.Link("Go back to lobby", atmPath+":lobby?user="+user))
		return
	}
	if game.state != gameContinue {
		rend.Write(renderEndGame(game, me.(*UserData), user, state) + nl)
		return
	}
	rend.Write(renderTeamSwap(game, me.(*UserData).team, user, state) + nl)
	rend.Write(renderInventoryUse(me.(*UserData).inventory, user, state) + nl)
	rend.Write(renderPokemon1v1(game, user, state) + nl)
	rend.Write(md.Link("Give up and go back to lobby", atmPath+":lobby?user="+user))
	return
}
