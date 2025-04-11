package pokemon

import (
	"errors"
	"strconv"
	"strings"

	"gno.land/p/leo/randomizer"
	"gno.land/p/leo/renderer"
	"gno.land/p/leo/utils"
)

var (
	rand = randomizer.NewRandomizer()
	nl   = "\n\n"
)

func getUrlFromGame(game Game) string {
	path := "?started=1&me=" + strconv.Itoa(game.myId) +
		"&other=" + strconv.Itoa(game.otherId) +
		"&myPv=" + strconv.Itoa(game.me.pv) +
		"&otherPv=" + strconv.Itoa(game.other.pv) +
		"&myPc="
	for id, attack := range game.me.attacks {
		if id != 0 {
			path += ","
		}
		path += strconv.Itoa(attack.pc)
	}
	path += "&otherPc="
	for id, attack := range game.other.attacks {
		if id != 0 {
			path += ","
		}
		path += strconv.Itoa(attack.pc)
	}
	return path
}

func getGameFromUrl(params map[string]string) (Game, error) {
	me, errMe := strconv.Atoi(params["me"])
	other, errOther := strconv.Atoi(params["other"])

	myPv, errMyPv := strconv.Atoi(params["myPv"])
	otherPv, errOtherPv := strconv.Atoi(params["otherPv"])

	myPc := strings.Split(params["myPc"], ",")
	otherPc := strings.Split(params["otherPc"], ",")
	if errMe != nil {
		return Game{}, errMe
	}
	if errOther != nil {
		return Game{}, errOther
	}
	if errMyPv != nil {
		return Game{}, errMyPv
	}
	if errOtherPv != nil {
		return Game{}, errOtherPv
	}
	if me >= len(pokemonList) {
		return Game{}, errors.New("game bad formatted : my id is bad")
	}
	if other >= len(pokemonList) {
		return Game{}, errors.New("game bad formatted : other id is bad")
	}
	if len(myPc) != len(pokemonList[me].attacks) {
		return Game{}, errors.New("game bad formatted : len of my pc missmatch")
	}
	if len(otherPc) != len(pokemonList[other].attacks) {
		return Game{}, errors.New("game bad formatted : len of other pc missmatch")
	}

	myPokemon := pokemonList[me].deepCopy()
	otherPokemon := pokemonList[other].deepCopy()
	myPokemon.pv = myPv
	otherPokemon.pv = otherPv

	for id, strPc := range myPc {
		pc, err := strconv.Atoi(strPc)
		if err != nil {
			return Game{}, errors.New("game my pc bad formatted")
		}
		myPokemon.attacks[id].pc = pc
	}
	for id, strPc := range otherPc {
		pc, err := strconv.Atoi(strPc)
		if err != nil {
			return Game{}, errors.New("game other pc bad formatted")
		}
		otherPokemon.attacks[id].pc = pc
	}
	return Game{me, myPokemon, other, otherPokemon}, nil
}

func renderLobbyPokemon(id int, pokemon Pokemon) (rend string) {
	rend += renderer.Header(3, pokemon.name)

	otherId := rand.Next() % len(pokemonList)
	path := getUrlFromGame(Game{myId: id, me: pokemon, otherId: otherId, other: pokemonList[otherId]})
	rend += renderer.Link(renderer.Picture(pokemon.name, pokemon.assetPath), path)
	rend += renderer.Bold(strconv.Itoa(pokemon.pv)) + " Pv\n"

	attackTab := [][]string{{"Attack"},
		{renderer.Bold("Damage")},
		{renderer.Bold("PC")}}
	for _, attack := range pokemon.attacks {
		attackTab[0] = append(attackTab[0], attack.name)
		attackTab[1] = append(attackTab[1], strconv.Itoa(attack.damage))
		attackTab[2] = append(attackTab[2], strconv.Itoa(attack.pc))
	}
	rend += renderer.Table(attackTab)
	return
}

func renderPokemonLobby() (rend string) {
	rend += renderer.Header(2, "Choose a Pokémon :")
	for id, pokemon := range pokemonList {
		rend += renderLobbyPokemon(id, pokemon)
	}
	return
}

func renderGamePokemon(id int, pokemon Pokemon) (rend string) {
	rend += renderer.Header(3, pokemon.name)

	otherId := rand.Next() % len(pokemonList)
	path := getUrlFromGame(Game{myId: id, me: pokemon, otherId: otherId, other: pokemonList[otherId]})
	rend += renderer.Link(renderer.Picture(pokemon.name, pokemon.assetPath), path)
	rend += renderer.Bold(strconv.Itoa(pokemon.pv)) + " Pv\n"

	attackTab := [][]string{{"Attack"},
		{renderer.Bold("Damage")},
		{renderer.Bold("PC")}}
	for _, attack := range pokemon.attacks {
		attackTab[0] = append(attackTab[0], attack.name)
		attackTab[1] = append(attackTab[1], strconv.Itoa(attack.damage))
		attackTab[2] = append(attackTab[2], strconv.Itoa(attack.pc))
	}
	rend += renderer.Table(attackTab)
	return
}

func renderGame(game Game) (rend string) {
	rend += renderer.Header(2, "Fight !")
	rend += renderer.Picture(game.me.name, game.me.assetPath) + nl

	myData := [][]string{
		{"Me", "PV"},
		{game.me.name, strconv.Itoa(game.me.pv) + " / " + strconv.Itoa(pokemonList[game.myId].pv)},
	}
	rend += renderer.Table(myData) + nl

	attackTab := [][]string{{"Attack"},
		{renderer.Bold("Damage")},
		{renderer.Bold("PC")},
		{""}}
	for idx, attack := range game.me.attacks {
		attackTab[0] = append(attackTab[0], attack.name)
		attackTab[1] = append(attackTab[1], strconv.Itoa(attack.damage))
		attackTab[2] = append(attackTab[2], strconv.Itoa(attack.pc)+" / "+strconv.Itoa(pokemonList[game.myId].attacks[idx].pc))

		otherAttack := rand.Next() % len(game.other.attacks)

		for _, attack := range game.other.attacks {
			if attack.pc > 0 {
				break
			}
			otherAttack = -1
		}

		gameCpy := game.deepCopy()
		if otherAttack != -1 {
			for {
				if game.other.attacks[otherAttack].pc > 0 {
					break
				}
				otherAttack = rand.Next() % len(game.other.attacks)
			}
			gameCpy.me.pv -= gameCpy.other.attacks[otherAttack].damage
			gameCpy.other.attacks[otherAttack].pc -= 1
		}

		if gameCpy.me.attacks[idx].pc > 0 {
			gameCpy.other.pv -= gameCpy.me.attacks[idx].damage
			gameCpy.me.attacks[idx].pc -= 1

			attackTab[3] = append(attackTab[3], renderer.Link(renderer.Picture("Use", "https://minecraft.wiki/images/Diamond_Sword_JE3_BE3.png"), getUrlFromGame(gameCpy)))
		} else {
			attackTab[3] = append(attackTab[3], renderer.Picture("No PC", "https://minecraft.wiki/images/archive/20150911135457%21Barrier_%28held%29_JE2_BE2.png"))
		}
	}
	rend += renderer.Table(attackTab)

	enemyData := [][]string{
		{"Ennemy", "PV"},
		{game.other.name, strconv.Itoa(game.other.pv) + " / " + strconv.Itoa(pokemonList[game.otherId].pv)},
	}
	rend += renderer.Table(enemyData) + nl

	rend += renderer.Picture(game.other.name, game.other.assetPath) + nl
	return
}

func endGame(win bool) (rend string) {
	if win {
		rend += renderer.Header(1, "WIN !")
	} else {
		rend += renderer.Header(1, "Loose...")
	}
	rend += renderer.Header(1, renderer.Link("Play Again ?", "?"))
	return
}

type RenderPage string

const (
	pokemonLobby        RenderPage = "lobby"
	pokemonLogin        RenderPage = "login"
	pokemonTeamSelector RenderPage = "team"
	pokemonFight        RenderPage = "fight"
	pokemonShop         RenderPage = "shop"
)

var (
	pokemonPath = map[RenderPage]func(map[string]string) string{
		pokemonLobby:        renderPokemonLogin,
		pokemonLogin:        renderPokemonLogin,
		pokemonTeamSelector: renderPokemonLogin,
		pokemonFight:        renderPokemonLogin,
		pokemonShop:         renderPokemonLogin,
	}
)

func Render(path string) (rend string) {
	rend += renderer.Header(1, "Pokémon")
	path, params := utils.ParsePath(path)

	if pathRenderer := pokemonPath[RenderPage(path)]; pathRenderer != nil && params["user"] != "" {
		rend += pathRenderer(params)
	} else {
		rend += renderPokemonLogin(params)
	}
	return
}
