package home

func r3() string {
	return `# Greps' notable contributions

My main contributions until the gno.land beta launch are listed below; most aren't in the monorepo, note.

### Port Joeson from coffeescript to golang

Worked on this from june 2022 until january 2023. Bounty [applied for](https://github.com/gnolang/bounties-old/issues/33) on Feb 2, 2023. 

Here is the port I did in Go: [grepsuzette/joeson](https://github.com/grepsuzette/joeson/).

    4. Port JOESON to Go
    github.com/jaekwon/joescript
    The intent is to create an independent left-recursive PEG parser for Gno.
    Optional: port Joescript or Javascript.
    1000 ATOMs from @jaekwon
    More GNOTs than from #3.

There have been many examples posted, including a minimal [LISP REPL](https://github.com/grepsuzette/joeson/tree/master/examples/lisp-repl) and a theorical [study on precedence](https://github.com/grepsuzette/joeson/blob/master/examples/precedence/precedence_test.go) (precedence is often problematic with PEG parsers, this allowed to find a solution, used in the next part). 

### GNO grammar - partial

In summer 2023, started to port the GNO grammar using Joeson (since there was no news about joeson, so this was an attempt to demonstrate it worked). Grammar was posted in [PR 1156](https://github.com/gnolang/gno/pull/1156). There are only 3 files, they are quite dense:

1. [joeson_test.go](https://github.com/grepsuzette/gno/blob/joeson/gnovm/pkg/gnolang/joeson_test.go)
1. [joeson_rules.go](https://github.com/grepsuzette/gno/blob/joeson/gnovm/pkg/gnolang/joeson_rules.go)
1. [joeson_f.go](https://github.com/grepsuzette/gno/blob/joeson/gnovm/pkg/gnolang/joeson_f.go)

### gnAsteroid

![asteroid](https://raw.githubusercontent.com/grepsuzette/gfx/master/asteroid160.png)

**gnAsteroid** is an asteroid creation-kit, it was started around the time the joeson port was published, but didn't have a name back then. 

Asteroids orbit gno.land, it's the same blockchain, but different frontend,
themable, working with wiki-like markdown files (enabling realms from gno.land
to be rendered there).

* [asteroid 0](https://gnAsteroid.com) - asteroid explaining what it is, containing instructions, to use, deploy on [Akash](https://gnasteroid.com/publishing/akash.md), [Vercel](https://gnasteroid.com/publishing/vercel.md).
* [greps' asteroid](https://greps.gnAsteroid.com)
* [gnAsteroid](https://github.com/gnAsteroid/gnAsteroid) - The github for gnAsteroid.

### Research with markdown and gnoweb, mini-games, experiments (summer-oct 2024)

A series of experiments with gnoweb 1.0 lead from the summer 2024, to try to advocate for keeping html
and css enabled in gnoweb, or at least to try to determine what we could
potentially miss without. Gnoweb1.0, markdown, html, css, js-less.

Note those still work with [gnAsteroid](https://gnAsteroid.com), or with gnoweb
running with the -web-html switch. As of now they are rendered through an
asteroid.

| 1                                                                                   | 2                                                                                     |
| :-------------------:                                                               | :-------------------------:                                                           |
| ![parrot](https://raw.githubusercontent.com/grepsuzette/gfx/master/parrot160.png)   | ![octopus](https://raw.githubusercontent.com/grepsuzette/gfx/master/octopus160.png)   |
| [tic-tac-toe](https://greps.gnAsteroid.com/r/grepsuzette/pr2554/v6/games/tictactoe) | [minesweeper](https://greps.gnAsteroid.com/r/grepsuzette/pr2554/v6/games/minesweeper) |

Check the [other experiments here](/conjects/gnoweb.md).

![octopus](https://raw.githubusercontent.com/grepsuzette/gfx/master/screen-minesweeper390.png)

### Tendermint vuln retrospective (2023)

Also worked on an anthology of publicly knowned vulnerabilities that affected Tendermint. 

* [Cosmos-sdk vulnerability retrospective](https://github.com/gnolang/gno/issues/587)
* found most vulns were not affecting our Tendermint version, however:
* [demonstrated vulnerability to BSC 2022-10-07 hack](https://github.com/gnolang/gno/pull/583)
* [proposed fix to vuln to BSC 2022-10-07 hack (merged)](https://github.com/gnolang/gno/pull/584)
* not all of them were tested, as I was hoping some more feedback before to continue.

There is also a small [GNO mail](https://github.com/gnolang/gno/pull/641) which got no UI is discussed in [one of my articles](https://greps.gnasteroid.com/articles/encryptedmail.md).

Thanks for reading!
`
}
