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
	nl = "\n\n"
)

type Attack struct {
	name   string
	damage int
	pc     int
}

type Pokemon struct {
	name      string
	assetPath string
	pv        int
	attacks   []Attack
}

type Game struct {
	myId    int
	me      Pokemon
	otherId int
	other   Pokemon
}

var (
	rand     = randomizer.NewRandomizer()
	pokemons = []Pokemon{
		{
			"Pikachu",
			"https://www.pokemon.com/static-assets/content-assets/cms2/img/pokedex/full/025.png",
			60,
			[]Attack{
				{"Thunderbolt", 50, 2},
				{"Tackle", 10, 10},
				{"Iron Tail", 30, 5},
			},
		},
		{
			"Bulbasaur",
			"https://www.pokemon.com/static-assets/content-assets/cms2/img/pokedex/full/001.png",
			80,
			[]Attack{
				{"Razor Leaf", 25, 6},
				{"False Swipe", 15, 10},
				{"Vine Whip", 45, 2},
			},
		},
		{
			"Charmander",
			"https://www.pokemon.com/static-assets/content-assets/cms2/img/pokedex/full/004.png",
			70,
			[]Attack{
				{"Flamethrower", 40, 3},
				{"Scratch", 10, 15},
				{"Fire Spin", 35, 4},
			},
		},
		{
			"Squirtle",
			"https://www.pokemon.com/static-assets/content-assets/cms2/img/pokedex/full/007.png",
			75,
			[]Attack{
				{"Water Gun", 30, 6},
				{"Tackle", 10, 12},
				{"Bubble Beam", 40, 3},
			},
		},
		{
			"Eevee",
			"https://www.pokemon.com/static-assets/content-assets/cms2/img/pokedex/full/133.png",
			65,
			[]Attack{
				{"Quick Attack", 20, 10},
				{"Swift", 25, 8},
				{"Bite", 30, 5},
			},
		},
		{
			"Jigglypuff",
			"https://www.pokemon.com/static-assets/content-assets/cms2/img/pokedex/full/039.png",
			85,
			[]Attack{
				{"Sing", 0, 5},
				{"Pound", 15, 12},
				{"Double Slap", 25, 6},
			},
		},
	}
)

func (p Pokemon) deepCopy() Pokemon {
	copiedAttacks := make([]Attack, len(p.attacks))
	copy(copiedAttacks, p.attacks)
	return Pokemon{
		name:      p.name,
		assetPath: p.assetPath,
		pv:        p.pv,
		attacks:   copiedAttacks,
	}
}

func (g Game) deepCopy() Game {
	return Game{
		myId:    g.myId,
		me:      g.me.deepCopy(),
		otherId: g.otherId,
		other:   g.other.deepCopy(),
	}
}

func getUrlFromGame(game Game) string {
	path := "?started=1&me=" + strconv.Itoa(game.myId) + "&other=" + strconv.Itoa(game.otherId) +
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
	if me >= len(pokemons) {
		return Game{}, errors.New("game bad formatted : my id is bad")
	}
	if other >= len(pokemons) {
		return Game{}, errors.New("game bad formatted : other id is bad")
	}
	if len(myPc) != len(pokemons[me].attacks) {
		return Game{}, errors.New("game bad formatted : len of my pc missmatch")
	}
	if len(otherPc) != len(pokemons[other].attacks) {
		return Game{}, errors.New("game bad formatted : len of other pc missmatch")
	}

	myPokemon := pokemons[me].deepCopy()
	otherPokemon := pokemons[other].deepCopy()
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

	otherId := rand.Next() % len(pokemons)
	path := getUrlFromGame(Game{myId: id, me: pokemon, otherId: otherId, other: pokemons[otherId]})
	rend += renderer.Link(renderer.Picture(pokemon.name, pokemon.assetPath), path)
	rend += renderer.Bold(strconv.Itoa(pokemon.pv)) + " Pv\n"

	attackTab := [][]string{[]string{"Attack"},
		[]string{renderer.Bold("Damage")},
		[]string{renderer.Bold("PC")}}
	for _, attack := range pokemon.attacks {
		attackTab[0] = append(attackTab[0], attack.name)
		attackTab[1] = append(attackTab[1], strconv.Itoa(attack.damage))
		attackTab[2] = append(attackTab[2], strconv.Itoa(attack.pc))
	}
	rend += renderer.Table(attackTab)
	return
}

func renderPokemonLobby() (rend string) {
	rend += renderer.Header(2, "Choose a pokemon :")
	for id, pokemon := range pokemons {
		rend += renderLobbyPokemon(id, pokemon)
	}
	return
}

func renderGamePokemon(id int, pokemon Pokemon) (rend string) {
	rend += renderer.Header(3, pokemon.name)

	otherId := rand.Next() % len(pokemons)
	path := getUrlFromGame(Game{myId: id, me: pokemon, otherId: otherId, other: pokemons[otherId]})
	rend += renderer.Link(renderer.Picture(pokemon.name, pokemon.assetPath), path)
	rend += renderer.Bold(strconv.Itoa(pokemon.pv)) + " Pv\n"

	attackTab := [][]string{[]string{"Attack"},
		[]string{renderer.Bold("Damage")},
		[]string{renderer.Bold("PC")}}
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
		[]string{"Me", "PV"},
		[]string{game.me.name, strconv.Itoa(game.me.pv) + " / " + strconv.Itoa(pokemons[game.myId].pv)},
	}
	rend += renderer.Table(myData) + nl

	attackTab := [][]string{[]string{"Attack"},
		[]string{renderer.Bold("Damage")},
		[]string{renderer.Bold("PC")},
		[]string{""}}
	for idx, attack := range game.me.attacks {
		attackTab[0] = append(attackTab[0], attack.name)
		attackTab[1] = append(attackTab[1], strconv.Itoa(attack.damage))
		attackTab[2] = append(attackTab[2], strconv.Itoa(attack.pc)+" / "+strconv.Itoa(pokemons[game.myId].attacks[idx].pc))

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
		[]string{"Ennemy", "PV"},
		[]string{game.other.name, strconv.Itoa(game.other.pv) + " / " + strconv.Itoa(pokemons[game.otherId].pv)},
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

func Render(path string) (rend string) {
	params := utils.ParseParams(path)
	gameState := params["started"]
	game, err := getGameFromUrl(params)
	rend += renderer.Header(1, "Pokemon")

	if gameState == "1" && err == nil {
		if game.me.pv <= 0 || game.other.pv <= 0 {
			rend += endGame(game.me.pv > game.other.pv)
			return
		}
		rend += renderGame(game)
	} else {
		if gameState == "1" {
			rend += err.Error() + nl
		}
		rend += renderPokemonLobby()
	}
	return
}
