package pokemon

import (
	"sort"
	"std"

	"gno.land/p/moul/md"
)

type playerEntry struct {
	adress     std.Address
	pokedollar uint
}

type playerList []playerEntry

func (p playerList) Len() int {
	return len(p)
}

func (p playerList) Less(i, j int) bool {
	return p[i].pokedollar > p[j].pokedollar
}

func (p playerList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func renderPokemonRank(params map[string]string) (rend string) {
	rend += md.H2("Ranking")
	membersAddress := pokemonDAO.GetMembers()

	var players playerList
	for i, u := range pokemonDAO.userData {
		players = append(players, playerEntry{
			adress:     membersAddress[i],
			pokedollar: u.pokedollar,
		})
	}

	sort.Sort(players)

	var playerOrdered []string
	for _, u := range players {
		if u.adress == std.Address(params["user"]) {
			playerOrdered = append(playerOrdered, "-> You: "+string(u.adress)+" : "+renderPokedollar(int(u.pokedollar)))
		} else {
			playerOrdered = append(playerOrdered, string(u.adress)+" : "+renderPokedollar(int(u.pokedollar)))
		}
	}
	rend += md.OrderedList(playerOrdered) + nl
	rend += md.Link("Go back to lobby", atm_path+":lobby?user="+params["user"])
	return
}
