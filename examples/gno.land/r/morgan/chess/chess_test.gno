package chess

import (
	"chain"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"gno.land/p/morgan/chess"
	"gno.land/p/morgan/chess/glicko2"
	"gno.land/p/nt/avl/v0"
)

func cleanup() {
	gameStore = avl.Tree{}
	gameIDCounter = 0
	user2Games = avl.Tree{}
	playerStore = avl.Tree{}
	leaderboard = [CategoryMax]leaderboardType{}
	lobby = [tcLobbyMax][]lobbyPlayer{}
	lobbyPlayer2Game = avl.Tree{}
	playerRatings = [CategoryMax][]*glicko2.PlayerRating{}
}

func TestNewGame(cur realm, t *testing.T) {
	cleanup()

	g := xNewGame(cur, chain.PackageAddress("xx").String(), 0, 0)
	println(g)
}

const (
	white address = "g1white"
	black address = "g1black"
)

/*
syntax:

	[<command> ][#[!][<buf>] <checker>]

command is executed; result of command is stored in buffer.
the test is split in lines. other white space is ignored (strings.Fields).

<buf>: all commands below will generally store a string result value in the
buffer "result", which is the default and thus may be omitted.
if the command panics, the panic value is stored in the buffer "panic".
(if it doesn't, buffer panic is set to an empty string).
if following a command there is no checker on the #panic buffer, the line
"#!panic empty" is implicitly added.
if <buf> is preceded by ! (e.g. "#!panic empty"), then if the checker fails,
processing is stopped on that line.

<command>:

	newgame [<white> <black> [<seconds> [<increment>]]]
		stores game ID in buffer #id.
		<white> and <black> are two addresses. if they are not passed, <white>
		assumes value "white" and <black> "black"
	move <player> <lan_move>
		lan_move is in the same format as Move.String.
		retrieves game id from #id.
	draw <player>
	drawoffer <player>
	abort <player>
	timeout <player>
		(ClaimTimeout)
	resign <player>
	game [<id>]
		if not given, id is retrieved from buffer #id.
	player <player>
	name <predicate>
		sets the name of the test to predicate.
	copy <dst> [<src>]
		copies buffer src to buffer dst.
		if src not specified, assumed result.
		(don't specify the #; ie: copy oldresult result)
	sleep <seconds>
		sleep for the given amount of seconds (float).

NOTE: for all values of <player>, including <white> and <black> in newgame,
the addresses are passed prefixed by "g1", and the matching checkers should
expect this.

<checker>:

	empty
		the buffer should be empty.
	equal <predicate>
		predicate may start with #, which indicates a buffer.
	contains [<predicate>...]
		the buffer should contain all of the given predicates.
	containssp <predicate>
		the buffer should contain the given predicate, which contains spaces.
*/
var commandTests = [...]string{
	`	name NewGameNegativeIncrement
		newgame white black 10 -5 #panic containssp negative increment invalid
	`,
	`	name NewGameDouble
		newgame
		newgame #panic contains is not yet finished
	`,
	`	name NewGameWithSelf
		newgame white white #panic contains game with yourself
	`,
	// ColoursInvert within games played by two players
	`	name ColoursInvert
		newgame
		move white e2e4
		move black e7e5
		move white f1c4
		resign white
		newgame
		# contains "white":"g1black" "black":"g1white"
		#id equal 0000002
	`,
	// Otherwise, invert from p1's history.
	`	name ColoursInvert3p
		newgame p1 p2 #! contains "white":"g1p1" "black":"g1p2"
		move p1 e2e4
		abort p1
		newgame p1 p3 # contains "white":"g1p3" "black":"g1p1"
	`,
	`	name ScholarsMate
		newgame #id equal 0000001
		move white e2e4
		move black e7e5
		move white f1c4
		move black b8c6
		move white d1f3
		move black d7d6
		move white f3f7
		copy moveres
		game # equal #moveres
		# contains "state":"checkmated" "winner":"white"
		# containssp r1bqkbnr/ppp2Qpp/2np4/4p3/2B1P3/8/PPPP1PPP/RNB1K1NR b KQkq - 0 4
		player white
		# contains "address":"g1white" "position":0 "wins":1 "losses":0 "draws":0
		player black
		# contains "address":"g1black" "position":1 "wins":0 "losses":1 "draws":0
	`,
	`	name DrawByAgreement
		newgame
		move white e2e4
		move black e7e5
		move white f1c4
		move black b8c6
		copy moveres
		game # equal #moveres
		# contains "open" "concluder":null "draw_offerer":null
		drawoffer white
		# contains "open" "concluder":null "draw_offerer":"g1white"
		draw black
		# contains "drawn_by_agreement" "concluder":"g1black" "draw_offerer":"g1white"
	`,
	`	name AbortFirstMove
		newgame
		abort white # contains "winner":"none" "concluder":"g1white"
	`,

	`	name ThreefoldRepetition
		newgame

		move white g1f3
		move black g8f6
		move white f3g1
		move black f6g8

		move white g1f3
		move black g8f6
		move white f3g1
		move black f6g8

		draw black # contains "winner":"draw" "concluder":"g1black"
		# contains "state":"drawn_3_fold"
	`,
	`	name FivefoldRepetition
		newgame

		move white g1f3
		move black g8f6
		move white f3g1
		move black f6g8

		move white g1f3
		move black g8f6
		move white f3g1
		move black f6g8

		move white g1f3
		move black g8f6
		move white f3g1
		move black f6g8

		move white g1f3
		move black g8f6
		move white f3g1
		move black f6g8

		# contains "winner":"draw" "concluder":null "state":"drawn_5_fold"

		move white g1f3 #panic contains game is already finished
	`,
	`	name TimeoutAborted
		newgame white black 3
		move white e2e4 #! contains "state":"open"
		sleep 31
		move black e7e5
		game
		# contains e2e4
		# contains "aborted"
		# contains "concluder":"g1black"
	`,
	`	name TimeoutAbandoned
		newgame white black 1
		move white e2e4
		move black e7e5
		sleep 61
		timeout black
		# contains "state":"timeout" "winner":"black"
	`,
}

func TestCommands(t *testing.T) {
	for _, command := range commandTests {
		runCommandTest(t, command)
	}
}

// testCommandRunner is used to represent the single testCommand types below.
// the Run function is used to execute the actual command, after parsing.
//
// This could have been implemented with simple closures generated within the
// parser, however it's not because of this issue:
// https://github.com/gnolang/gno/issues/1135
type testCommandRunner interface {
	Run(t *testing.T, bufs map[string]string)
}

// testCommandChecker is a wrapper for a runner which is performing a check.
// This is marked in order not to wrap the calls to them as a panic.
type testCommandChecker struct{ testCommandRunner }

func (testCommandChecker) Checker() {}

type testCommandFunc func(t *testing.T, bufs map[string]string)

func (tc testCommandFunc) Run(t *testing.T, bufs map[string]string) { tc(t, bufs) }

// testCommandColorID represents a testCommand, which uses a function of the
// form func(gameID string) string (hence ID), and that takes as the first
// parameter a <player> which will be the caller.
type testCommandColorID struct {
	fn   func(realm, string) string
	addr address
}

func newTestCommandColorID(fn func(realm, string) string, s string, addr string) testCommandRunner {
	return &testCommandColorID{fn, address("g1" + addr)}
}

func (tc *testCommandColorID) Run(t *testing.T, bufs map[string]string) {
	testing.SetRealm(testing.NewUserRealm(tc.addr))
	bufs["result"] = tc.fn(cross, bufs["id"])
}

type testCommandNewGame struct {
	w, b          address
	seconds, incr int
}

func (tc *testCommandNewGame) Run(t *testing.T, bufs map[string]string) {
	testing.SetRealm(testing.NewUserRealm(tc.w))
	res := xNewGame(cross, string(tc.b), tc.seconds, tc.incr)
	bufs["result"] = res

	const idMagicString = `"id":"`
	idx := strings.Index(res, idMagicString)
	if idx < 0 {
		panic("id not found")
	}
	id := res[idx+len(idMagicString):]
	id = id[:strings.IndexByte(id, '"')]
	bufs["id"] = id
}

type testCommandMove struct {
	addr      address
	from, to  string
	promotion chess.Piece
}

func (tc *testCommandMove) Run(t *testing.T, bufs map[string]string) {
	testing.SetRealm(testing.NewUserRealm(tc.addr))
	bufs["result"] = MakeMove(cross, bufs["id"], tc.from, tc.to, tc.promotion)
}

type testCommandGame struct {
	idWanted string
}

func (tc *testCommandGame) Run(t *testing.T, bufs map[string]string) {
	idl := tc.idWanted
	if idl == "" {
		idl = bufs["id"]
	}
	bufs["result"] = GetGame(idl)
}

type testCommandPlayer struct {
	addr string
}

func (tc *testCommandPlayer) Run(t *testing.T, bufs map[string]string) {
	bufs["result"] = GetPlayer(tc.addr)
}

type testCommandCopy struct {
	dst, src string
}

func (tc *testCommandCopy) Run(t *testing.T, bufs map[string]string) {
	bufs[tc.dst] = bufs[tc.src]
}

type testCommandSleep struct {
	dur time.Duration
}

func (tc *testCommandSleep) Run(t *testing.T, bufs map[string]string) {
	os.Sleep(tc.dur)
}

type testChecker struct {
	fn    func(t *testing.T, bufs map[string]string, tc *testChecker)
	tf    func(*testing.T, string, ...interface{})
	bufp  string
	preds []string
}

func (*testChecker) Checker() {}
func (tc *testChecker) Run(t *testing.T, bufs map[string]string) {
	tc.fn(t, bufs, tc)
}

func parseCommandTest(t *testing.T, command string) (funcs []testCommandRunner, testName string) {
	lines := strings.Split(command, "\n")
	atoi := func(s string) int {
		n, err := strconv.Atoi(s)
		checkErr(err)
		return n
	}
	// used to detect whether to auto-add a panic checker
	var hasPanicChecker bool
	panicChecker := func(lineNum int, testName string) testCommandRunner {
		return testCommandChecker{testCommandFunc(
			func(t *testing.T, bufs map[string]string) {
				if bufs["panic"] != "" {
					t.Fatalf("%s:%d: buffer \"panic\" is not empty (%q)", testName, lineNum, bufs["panic"])
				}
			},
		)}
	}

	for lineNum, line := range lines {
		flds := strings.Fields(line)
		if len(flds) == 0 {
			continue
		}
		command, checker := flds, ([]string)(nil)
		for idx, fld := range flds {
			if strings.HasPrefix(fld, "#") {
				command, checker = flds[:idx], flds[idx:]
				break
			}
		}
		var cmd string
		if len(command) > 0 {
			cmd = command[0]

			// there is a new command; if hasPanicChecker == false,
			// it means the previous command did not have a panic checker.
			// add it.
			if !hasPanicChecker && len(funcs) > 0 {
				// no lineNum+1 because it was the previous line
				funcs = append(funcs, panicChecker(lineNum, testName))
			}
		}
		switch cmd {
		case "": // move on
		case "newgame":
			w, b := white, black
			var seconds, incr int
			switch len(command) {
			case 1:
			case 5:
				incr = atoi(command[4])
				fallthrough
			case 4:
				seconds = atoi(command[3])
				fallthrough
			case 3:
				w, b = address("g1"+command[1]), address("g1"+command[2])
			default:
				panic("invalid newgame command " + line)
			}
			funcs = append(funcs,
				&testCommandNewGame{w, b, seconds, incr},
			)
		case "move":
			if len(command) != 3 {
				panic("invalid move command " + line)
			}
			if len(command[2]) < 4 || len(command[2]) > 5 {
				panic("invalid lan move " + command[2])
			}
			from, to := command[2][:2], command[2][2:4]
			var promotion chess.Piece
			if len(command[2]) == 5 {
				promotion = chess.PieceFromChar(command[2][4])
				if promotion == chess.PieceEmpty {
					panic("invalid piece for promotion: " + string(command[2][4]))
				}
			}
			funcs = append(funcs, &testCommandMove{
				addr:      address("g1" + command[1]),
				from:      from,
				to:        to,
				promotion: promotion,
			})
		case "abort":
			funcs = append(funcs, newTestCommandColorID(Abort, "abort", command[1]))
		case "draw":
			funcs = append(funcs, newTestCommandColorID(Draw, "draw", command[1]))
		case "drawoffer":
			funcs = append(funcs, newTestCommandColorID(DrawOffer, "drawoffer", command[1]))
		case "timeout":
			funcs = append(funcs, newTestCommandColorID(ClaimTimeout, "timeout", command[1]))
		case "resign":
			funcs = append(funcs, newTestCommandColorID(Resign, "resign", command[1]))
		case "game":
			if len(command) > 2 {
				panic("invalid game command " + line)
			}
			tc := &testCommandGame{}
			if len(command) == 2 {
				tc.idWanted = command[1]
			}
			funcs = append(funcs, tc)
		case "player":
			if len(command) != 2 {
				panic("invalid player command " + line)
			}
			funcs = append(funcs, &testCommandPlayer{"g1" + command[1]})
		case "name":
			testName = strings.Join(command[1:], " ")
		case "copy":
			if len(command) > 3 || len(command) < 2 {
				panic("invalid copy command " + line)
			}
			tc := &testCommandCopy{dst: command[1], src: "result"}
			if len(command) == 3 {
				tc.src = command[2]
			}
			funcs = append(funcs, tc)
		case "sleep":
			if len(command) != 2 {
				panic("invalid sleep command " + line)
			}
			funcs = append(funcs, &testCommandSleep{
				time.Duration(atoi(command[1])) * time.Second,
			})
		default:
			panic("invalid command " + cmd)
		}

		if len(checker) == 0 {
			continue
		}
		if len(checker) == 1 {
			panic("no checker specified " + line)
		}

		bufp := checker[0]
		useFatal := false
		if len(bufp) > 1 && bufp[1] == '!' {
			bufp = bufp[2:]
			useFatal = true
		} else {
			bufp = bufp[1:]
		}
		if bufp == "" {
			bufp = "result"
		}
		if bufp == "panic" && !hasPanicChecker {
			hasPanicChecker = true
		}
		tf := func(ln int, testName string, useFatal bool) func(*testing.T, string, ...interface{}) {
			return func(t *testing.T, s string, v ...interface{}) {
				fn := t.Errorf
				if useFatal {
					fn = t.Fatalf
				}
				fn("%s:%d: "+s, append([]interface{}{testName, ln}, v...)...)
			}
		}(lineNum+1, testName, useFatal)

		switch checker[1] {
		case "empty":
			if len(checker) != 2 {
				panic("invalid empty checker " + line)
			}
			funcs = append(funcs, &testChecker{
				fn: func(t *testing.T, bufs map[string]string, tc *testChecker) {
					if bufs[tc.bufp] != "" {
						tc.tf(t, "buffer %q is not empty (%v)", tc.bufp, bufs[tc.bufp])
					}
				},
				tf:   tf,
				bufp: bufp,
			})
		case "equal":
			pred := strings.Join(checker[2:], " ")
			funcs = append(funcs, &testChecker{
				fn: func(t *testing.T, bufs map[string]string, tc *testChecker) {
					exp := tc.preds[0]
					if exp[0] == '#' {
						exp = bufs[exp[1:]]
					}
					if bufs[tc.bufp] != exp {
						tc.tf(t, "buffer %q: want %v got %v", tc.bufp, exp, bufs[tc.bufp])
					}
				},
				tf:    tf,
				bufp:  bufp,
				preds: []string{pred},
			})
		case "contains":
			preds := checker[2:]
			if len(preds) == 0 {
				break
			}
			funcs = append(funcs, &testChecker{
				fn: func(t *testing.T, bufs map[string]string, tc *testChecker) {
					for _, pred := range tc.preds {
						if !strings.Contains(bufs[tc.bufp], pred) {
							tc.tf(t, "buffer %q: %v does not contain %v", tc.bufp, bufs[tc.bufp], pred)
						}
					}
				},
				tf:    tf,
				bufp:  bufp,
				preds: preds,
			})
		case "containssp":
			pred := strings.Join(checker[2:], " ")
			if pred == "" {
				panic("invalid contanssp checker " + line)
			}
			funcs = append(funcs, &testChecker{
				fn: func(t *testing.T, bufs map[string]string, tc *testChecker) {
					if !strings.Contains(bufs[tc.bufp], tc.preds[0]) {
						tc.tf(t, "buffer %q: %v does not contain %v", tc.bufp, bufs[tc.bufp], tc.preds[0])
					}
				},
				tf:    tf,
				bufp:  bufp,
				preds: []string{pred},
			})
		default:
			panic("invalid checker " + checker[1])
		}
	}
	if !hasPanicChecker {
		funcs = append(funcs, panicChecker(len(lines), testName))
	}
	return
}

func runCommandTest(t *testing.T, command string) {
	funcs, testName := parseCommandTest(t, command)

	t.Run(testName, func(t *testing.T) {
		cleanup()
		bufs := make(map[string]string, 3)
		for _, f := range funcs {
			if _, ok := f.(interface{ Checker() }); ok {
				f.Run(t, bufs)
			} else {
				catchPanic(f, t, bufs)
			}
		}
	})
}

func catchPanic(tc testCommandRunner, t *testing.T, bufs map[string]string) {
	// XXX: should prefer testing.Recover, but see: https://github.com/gnolang/gno/issues/1650
	e := revive(func() { tc.Run(t, bufs) })
	if e == nil {
		bufs["panic"] = ""
		return
	}
	bufs["result"] = ""
	bufs["panic"] = fmt.Sprint(e)
}
