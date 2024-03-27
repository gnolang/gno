package game2048

import (
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/users"
)

func Render(path string) string {
	if path == "" {
		return renderTopPlayers()
	}
	player := getPlayer(users.AddressOrName(path).Resolve())
	if player == nil {
		return "404: player not found\n"
	}
	return renderPlayerView(player)
}

func renderTopPlayers() (res string) {
	topPlayers := leaderboard()
	for i, player := range topPlayers {
		res += ufmt.Sprintf("%d. %s - %d\n", i+1, player.Id.String(), player.Score)
	}

	return
}

func renderPlayerView(p *Player) (res string) {
	board := p.Board.Display()

	res += ufmt.Sprintf("# Board\n")
	res += ufmt.Sprintf("```\n%s\n```\n", board)
	res += ufmt.Sprintf("# Player Id: %s\n", p.Id.String())
	res += ufmt.Sprintf("# Score %d\n", p.Score)

	return
}
