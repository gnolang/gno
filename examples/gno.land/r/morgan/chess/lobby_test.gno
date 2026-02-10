package chess

import (
	"os"
	"testing"
	"time"

	"gno.land/p/morgan/chess/glicko2"
)

func TestLobbyJoin(t *testing.T) {
	cleanup()
	testing.SetRealm(testing.NewUserRealm(white))
	LobbyJoin(cross, 10*60, 5)
	os.Sleep(time.Second * 5)
	testing.SetRealm(testing.NewUserRealm(black))
	LobbyJoin(cross, 10*60, 5)
	res := LobbyGameFound(cross)
	if res == "null" {
		t.Errorf("LobbyGameFound is null")
	}
}

func sublobbyToIDs(pl []lobbyPlayer) []string {
	s := make([]string, len(pl))
	for idx, p := range pl {
		s[idx] = string(p.player.Address)
	}
	return s
}

func TestLobbyGameFound(t *testing.T) {
	check := func(checker ...func(t *testing.T)) func(t *testing.T) {
		return func(t *testing.T) {
			for _, ck := range checker {
				ck(t)
			}
		}
	}
	ids := func(ids ...address) func(t *testing.T) {
		return func(t *testing.T) {
			if len(ids) != len(lobby[0]) {
				t.Errorf("lobby doesn't match expected ids: lobby: %v, newIDs: %v", sublobbyToIDs(lobby[0]), ids)
				return
			}
			for idx, i := range ids {
				if pa := lobby[0][idx].player.Address; pa != i {
					t.Errorf("check pos %d: player id doesnt match (got %q want %q)", idx, pa, i)
				}
			}
		}
	}
	numGames := func(n int) func(t *testing.T) {
		return func(t *testing.T) {
			l := gameStore.Size()
			if l != n {
				t.Errorf("invalid gameStore size; want %d got %d", n, l)
			}
		}
	}

	type pl struct {
		id     address
		rating float64
		// use negative values here to indicate how many seconds in the past;
		// ie: joinedAt: -1, means player joined 1 second ago.
		joinedAt int
		seenAt   int
	}
	tt := []struct {
		name   string
		pre    []pl
		caller address
		check  func(t *testing.T)
	}{
		{
			"equalRating",
			[]pl{{"1", 1200, -1, -1}, {"2", 1200, 0, 0}},
			"1",
			check(ids(), numGames(1)),
		},
		{
			"minimumApart", // delta <= 25
			[]pl{{"1", 1200, 0, 0}, {"2", 1225, 0, 0}},
			"2",
			check(ids(), numGames(1)),
		},
		{
			"tooFarApart", // delta > 25
			[]pl{{"1", 1200, 0, 0}, {"2", 1230, 0, 0}},
			"2",
			check(ids("1", "2"), numGames(0)),
		},
		{
			"oldHighPriority",
			// kicked hasn't been seen in too long, so should not be considered.
			// 1 is active and has been looking for 30s, so it gets priority, even if 2-3 is
			// a closer match.
			[]pl{{"kicked", 1800, -60, -50}, {"1", 1900, -30, -10}, {"2", 1400, 0, 0}, {"3", 1420, 0, 0}},
			"3",
			check(ids("2"), numGames(1)),
		},
		{
			"oldHighPriority2",
			[]pl{{"comeback", 1800, -60, -50}, {"1", 1900, -30, -10}, {"2", 1400, 0, 0}, {"3", 1420, 0, 0}},
			// same as last one, except the player who was kicked last time, because
			// he's the caller, has their seenAt set back to the current time, so they're matched with 1.
			"comeback",
			check(ids("2", "3"), numGames(1)),
		},
		{
			"alone",
			[]pl{{"1", 1200, 0, 0}},
			"1",
			check(ids("1"), numGames(0)),
		},
		{
			"brackFail",
			[]pl{{"1", 1200, -4, -4}, {"2", 1450, -5, -5}},
			"1",
			check(ids("1", "2"), numGames(0)),
		},
		{
			"brackFail2",
			[]pl{{"1", 1200, -5, -5}, {"2", 1450, -4, -4}},
			"1",
			check(ids("1", "2"), numGames(0)),
		},
		{
			"brackSuccess",
			[]pl{{"1", 1200, -5, -5}, {"2", 1450, -5, -5}},
			"1",
			check(ids(), numGames(1)),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cleanup()
			now := time.Now()
			for _, p := range tc.pre {
				lobby[0] = append(lobby[0], lobbyPlayer{
					joinedAt: now.Add(time.Duration(p.joinedAt) * time.Second),
					seenAt:   now.Add(time.Duration(p.seenAt) * time.Second),
					player: &Player{
						Address: p.id,
						CategoryInfo: [CategoryMax]CategoryInfo{
							Blitz: {PlayerRating: &glicko2.PlayerRating{Rating: p.rating}},
						},
					},
				})
			}

			testing.SetRealm(testing.NewUserRealm(tc.caller))
			LobbyGameFound(cross)

			if tc.check != nil {
				tc.check(t)
			}
		})
	}
}

func TestLobbyJoin_HasOpenGames(t *testing.T) {
	cleanup()
	g := &Game{
		ID:    "123",
		White: white,
		Black: black,
		State: GameStateOpen,
	}
	gameStore.Set(g.ID, g)
	addToUser2Games(white, g)
	addToUser2Games(black, g)

	testing.SetRealm(testing.NewUserRealm(white))
	LobbyJoin(cross, 10*60, 5)
	if g.State != GameStateAborted {
		t.Errorf("state wrong: want %d got %d", GameStateAborted, g.State)
	}
	if g.Winner != WinnerNone {
		t.Errorf("winner wrong: want %q got %q", "none", g.Winner)
	}
}
