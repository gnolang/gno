package home

func Render(path string) string {
	switch path {
	case "3":
		return r3()
	case "2":
		return r2()
	default:
		return r1()
	}
}

const tripleBackquote = "```"
const art = `
                           (    )
                          (    )
                            )  )
                           (  (                  /\
                            (_)                 /  \  /\
                    ________[_]________      /\/    \/  \
           /\      /\        ______    \    /   /\/\  /\/\
          /  \    //_\       \    /\    \  /\/\/    \/    \
   /\    / /\/\  //___\       \__/  \    \/      +     '   
  /  \  /\/    \//_____\       \ |[]|     \    .    t     .
 /\/\/\/       //_______\       \|__|      \     p    e    
/      \      /XXXXXXXXXX\                  \      o    l  
        \    /_I_II  I__I_\__________________\  +    r    e
               I_I|  I__I_____[]_|_[]_____I      .     t   
               I_II  I__I_____[]_|_[]_____I          +    '
               I II__I  I     XXXXXXX     I    
            ~~~~~"   "~~~~~~~~~~~~~~~~~~~~~~~~   :*:*:*:*:*
`

func r1() string {
	return "# greps' (gn)home" +
		`
You've reached the terrestrial realms of Grepsuzette on gno.land. 

` + tripleBackquote + art + tripleBackquote + `

I am often on my [GNO asteroid](https://greps.gnAsteroid.com) too.

* Public address: g1fjh9y7ausp27dqsdq0qrcsnmgvwm6829v2au7d
* Contributor since summer 2022 ([notable contributions](/r/grepsuzette/home:3))
* You can try my games in GNO (they use gnoweb -html):
  * [tic-tac-toe](https://greps.gnasteroid.com/r/grepsuzette/pr2554/v6/games/tictactoe)
  * [minesweeper](https://greps.gnasteroid.com/r/grepsuzette/pr2554/v6/games/minesweeper)
`
}

func r2() string {
	return `A manual index, until there's an automated way:

* [home](home/): greps' home on gno.land
* [games](games/): series of games

I'm often on my [GNO asteroid][1] too.

[1]: https://greps.gnAsteroid.com
`
}
