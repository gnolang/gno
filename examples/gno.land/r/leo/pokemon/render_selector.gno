package pokemon

import (
	"std"

	"gno.land/p/leo/randomizer"
	"gno.land/p/leo/utils"
	"gno.land/p/moul/md"
	"gno.land/r/leon/hor"
)

type RenderPage string

const (
	pokemonLobby        RenderPage = "lobby"
	pokemonLogin        RenderPage = "login"
	pokemonTeamSelector RenderPage = "team"
	pokemonFight        RenderPage = "fight"
	pokemonShop         RenderPage = "shop"
	pokemonRank         RenderPage = "rank"
)

var (
	rand        = randomizer.NewRandomizer()
	nl          = "\n\n"
	atm_path    = std.CurrentRealm().PkgPath()[len(std.ChainDomain()):]
	pokemonPath = map[RenderPage]func(map[string]string) string{
		pokemonLogin:        renderPokemonLogin,
		pokemonLobby:        renderPokemonLobby,
		pokemonTeamSelector: renderPokemonTeam,
		pokemonFight:        renderPokemonFight,
		pokemonShop:         renderPokemonShop,
		pokemonRank:         renderPokemonRank,
	}
)

func init() {
	hor.Register("Leo's Pokémon", "Come play Pokémon on the Block Chain !")
}

func Render(path string) (rend string) {
	rend += md.H1("Pokémon")
	path, params := utils.ParsePath(path)

	if pathRenderer := pokemonPath[RenderPage(path)]; pathRenderer != nil && pokemonDAO.IsMember(std.Address(params["user"])) {
		rend += pathRenderer(params)
	} else {
		rend += renderPokemonLogin(params)
	}
	return
}
