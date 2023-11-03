package flippando

import (
	"strconv"
	"bytes"
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/flippandoserver"
)

type GameBoard struct {
	Board     []int64  `json:"board"`
	GameTiles []int64  `json:"gameTiles"`
	BoardSize int    `json:"boardSize"`
}

type Game struct {
	ID 			string `json:"id"`
	GameStatus 	string	`json:"gameStatus"`
	GameBoard 	[]int64	`json:"gameBoard"`
	SolvedGameBoard	[]int64	`json:"solvedGameBoard"`
	GameTiles	[]int64	`json:"gameTiles"`
	TileType	string	`json:"-"`
}

type FlippandoBasicNFT struct {
	name	string `json:"name"`
	version	string	`json:"version"`
	gameId	string`json:"gameId"`
	gameType string `json:"gameType"`
	svgData	string `json:"svgData"`
}

type nftResult struct {
	result	string `json:"result"`
	nftError	string	`json:"error"`
}

type inTransit struct {
	tokenId   int
	inTransit bool
}

// tokenId - int, inTransit, bool

// realm state
var (	
	gameStore     avl.Tree // string (game ID) -> *Game
	gameIDCounter uint64
	tokenIDCounter	uint64

	user2Games avl.Tree // std.Address -> []*Game
)

/*
// test functions
func TileTest(solvedBoard string, size int) string {
	solveBoardArray, err := ParseString(solvedBoard);
	if (err != nil){
		panic("error in parsing args")
	}

	testSVGData := flippandoserver.TestTiles(solveBoardArray, size);
	testBasicNFT := &FlippandoBasicNFT{
		name: "Flippando - the Game",
		version: "1.0.0",
		gameId: "testGameId",
		gameType: "testGameType",
		svgData: testSVGData,
	}
	testBasicNFTJSON, err := testBasicNFT.MarshalJSON()
	if err != nil {
		panic("error in marhasling testBasicNFT");
	}
	return string(testBasicNFTJSON)
}*/

// creates game object and initializes board
func StartGame(player std.Address, tileType string, boardSize int) string {
	
	games := GetUserGames(player)
	// Set up Game struct. Save in gameStore and user2games.
	gameIDCounter++
	board, gameTiles, boardSize := flippandoserver.CreateGameBoard(tileType, boardSize)
	
	// id is zero-padded to work well with avl's alphabetic order.
	id := zeroPad9(strconv.FormatUint(gameIDCounter, 10))
	g := &Game{
		ID: id,
		GameBoard: make([]int64, len(board)),
		SolvedGameBoard: make([]int64, len(board)),
		GameTiles: gameTiles,
		TileType: tileType,
		GameStatus: "initialized",
	}
	copy(g.GameBoard, board)
	copy(g.SolvedGameBoard, board)

	gameStore.Set(g.ID, g)
	addToUser2Games(player, g)
	// write the marshalJSON function for game
	// call marshall JSON and return the string
	newGameJSON, err := g.MarshalJSON()
	if err != nil {
		return "error";
	}
	return string(newGameJSON)
}


func FlipTiles(player std.Address, gameID string, positions string) (string) {
	// get the current game of the player
	games := GetUserGames(player)
	currentGame := games[0]
	currentGamePositions, err := ParseString(positions)
	
	// call real FlipTiles on flippandoserver, which does the game mechanics
	// and returns 2 []int64 boards: the GameBoard, and the SolvedGameBoard
	
	board, solvedBoard := flippandoserver.FlipTiles(currentGame.GameBoard, 
		currentGame.SolvedGameBoard, 
		currentGame.GameTiles, 
		currentGamePositions)
	currentGame.GameBoard = board
	currentGame.SolvedGameBoard = solvedBoard
	hasZeroValues := func(s []int64) bool { for _, v := range s { if v == 0 { return true } }; return false }(currentGame.SolvedGameBoard)
	if(!hasZeroValues){
		currentGame.GameStatus = "finished"
	}
	currentGameJSON, err := currentGame.MarshalJSON()
	if err != nil {
		return "error";
	}
	return string(currentGameJSON)
}

func CreateNFT(player std.Address, gameId string) (string) {
	result := ""
	nftError := ""
	// - generates a GRC721 token with the solved board SVG as a tokenURI
	basicNFT := grc721.NewBasicNFT("Flippando Basic Board", "FLPP")
	games := GetUserGames(player)
	// safety checks
	if(len(games) == 0){
		result = "{failedSafetyCheck: userHasGames}"
		nftError = "{error: user has zero games}"
	}
	currentGame := games[0]
	hasZeroValues := func(s []int64) bool { for _, v := range s { if v == 0 { return true } }; return false }(currentGame.SolvedGameBoard)
	if(hasZeroValues){
		result := "{failedSafetyCheck: gameCompletion}"
		nftError := "{error: game is not finished}"
	}
	gameSVGData, errSVG := GenerateColorSquareGrid(currentGame.SolvedGameBoard, 4)
	if(errSVG != nil){
		result := "{failedSafetyCheck: svgDataCorrupted}"
		nftError := "{error: SVG data not generated}"
	}
	flippandoBasicNFT := &FlippandoBasicNFT{
		name: "Flippando - the Game",
		version: "1.0.0",
		gameId: currentGame.ID,
		gameType: currentGame.TileType,
		svgData: gameSVGData,
	}
	flippandoBasicNFTJSON, err := flippandoBasicNFT.MarshalJSON()
	if err != nil {
		panic("error in marhasling flippandoBasicNFT");
	}
	mintResultError := basicNFT.SafeMint(player, grc721.TokenID(strconv.FormatUint(tokenIDCounter, 10)))
	if(mintResultError == nil){
		basicNFT.SetTokenURI(grc721.TokenID(strconv.FormatUint(tokenIDCounter, 10)), grc721.TokenURI(flippandoBasicNFTJSON))
		tokenIDCounter++
		
		result = string(flippandoBasicNFTJSON)
		nftError = "{error: false}"
	}
	res := &nftResult{
		result: result,
		nftError: nftError,
	}
	
	createNFTResult, err := res.MarshalJSON()
	if(err != nil){
		return "error"
	}
	return string(createNFTResult)
}

func GetUserNFTs(addr std.Address) string {
	tokenCount := grc721.BalanceOf(addr)
	tokenIDs := make([]int64, 0, tokenCount) 

	// Assuming there's a function to get the token ID by index for the given address
	for i := int64(0); i < tokenCount; i++ {
		owner, err := basicNFT.OwnerOf(i)
		if err != nil {
			if err == ErrInvalidTokenId {
				// do we need to do smth else here?
				continue
			} else {
				// resort to nil
				return nil, err
			}
		}
		
		// If the owner matches the address we're looking for, add the TokenID to the slice.
		if owner == addr {
			tokenIDs = append(tokenIDs, i)
		}
	}

	return tokenIDs
}


    
func GetUserGames(user std.Address) []*Game {
	val, exist := user2Games.Get(user.String())
	var games []*Game
	if !exist {
		return nil
	}
	return val.([]*Game)
}

func addToUser2Games(addr std.Address, game *Game) {
	var games []*Game
	v, ok := user2Games.Get(string(addr))
	if ok {
		games = v.([]*Game)
	}
	// game must be at top, because it is the latest ID
	games = append([]*Game{game}, games...)
	user2Games.Set(string(addr), games)
}

const zeroes = "000000000"

func zeroPad9(s string) string {
	n := 9 - len(s)
	if n < 0 {
		return s
	}
	return zeroes[:n] + s
}

func (g Game) MarshalJSON() (_ []byte, err error) {
	var b bytes.Buffer

	b.WriteByte('{')
	
	b.WriteString(`"id":"`)
	b.WriteString(g.ID)
	b.WriteString(`",`)

	b.WriteString(`"gameBoard":[`)
	b.WriteString(writeIntSlice(g.GameBoard))
	b.WriteString(`],`)

	b.WriteString(`"solvedGameBoard":[`)
	b.WriteString(writeIntSlice(g.SolvedGameBoard))
	b.WriteString(`],`)

	b.WriteString(`"gameTiles":[`)
	b.WriteString(writeIntSlice(g.GameTiles))
	b.WriteString(`],`)
	
	b.WriteString(`"gameStatus":"`)
	b.WriteString(g.GameStatus)
	b.WriteString(`"`)
	
	b.WriteByte('}')

	return b.Bytes(), nil
}

func (gb GameBoard) MarshalJSON() (_ []byte, err error) {
	var b bytes.Buffer

	b.WriteByte('{')
	
	b.WriteString(`"BoardSize":`)
	b.WriteString(strconv.Itoa(gb.BoardSize)) // converting int to string
	b.WriteString(`,`)
	
	
	b.WriteString(`"Board":[`)
	b.WriteString(writeIntSlice(gb.Board))
	b.WriteString(`],`)
	
	b.WriteString(`"GameTiles":[`)
	b.WriteString(writeIntSlice(gb.GameTiles))
	b.WriteString(`]`)
	
	b.WriteByte('}')

	return b.Bytes(), nil
}

func (fbn FlippandoBasicNFT) MarshalJSON() (_ []byte, err error) {
	
	var b bytes.Buffer

	b.WriteByte('{')
	
	b.WriteString(`"name":"`)
	b.WriteString(fbn.name)
	b.WriteString(`",`)

	b.WriteString(`"version":"`)
	b.WriteString(fbn.version)
	b.WriteString(`",`)

	b.WriteString(`"gameId":"`)
	b.WriteString(fbn.gameId)
	b.WriteString(`",`)

	b.WriteString(`"gameType":"`)
	b.WriteString(fbn.gameType)
	b.WriteString(`",`)
	
	b.WriteString(`"svgData":"`)
	b.WriteString(fbn.svgData) 
	b.WriteString(`"`)
	
	b.WriteByte('}')

	return b.Bytes(), nil
	
}

func (n nftResult)MarshalJSON() (_ []byte, err error) {
	
	var b bytes.Buffer

	b.WriteByte('{')

	b.WriteString(`"result":"`)
	b.WriteString(n.result)
	b.WriteString(`",`)

	b.WriteString(`"error":"`)
	b.WriteString(n.nftError)
	b.WriteString(`",`)

	b.WriteByte('}')

	return b.Bytes(), nil

}

func writeIntSlice(slice []int64) string {
    var b strings.Builder
    for i, v := range slice {
        if i > 0 {
            b.WriteString(",")
        }
        b.WriteString(strconv.FormatInt(v, 10))  // converting int64 to string
    }
    return b.String()
}


func ParseString(s string) ([]int64, error) {
	s = strings.Trim(s, "[]")
	parts := strings.Split(s, ",")
	ints := make([]int64, len(parts))
	for i, part := range parts {
		val, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		ints[i] = int64(val)
	}
	return ints, nil
}



func makeArt(owner string, buildingBlocks []int64) {
	// - checks if any of the tokenIds in the buildingBlocks[] are belonging to the user (not allowed)
	// - checks if any of the tokenIds in the buildingBlocks[] are in transit
	// - unlocks and sends the locked Flips in each used NFT
	// - transfers the NFTs to the new owner (the art creator)
	// - calls FlippandoBundler.bundleAssets() to generate a new GRC721 NFT with a standard tokenURI, and the tokenIds of the used NFT primitives
}

func ResetGame(gameId string) {
	//- utility function, nullifies the game[gameId]
}

// utility, private functions
func boardHasZeroValues(gameId string) {
	// - utility function, checks for existing zeros in the board (game is ongoing) // do we still need this?
}

func isValidGameLevel(boardSize int) {
	// - utility function, checks valid game levels
}

func isValidGameType(tileType string) {
	//   - utility function, checks valid tile types
}

func RenderHome() string {
	return "render path"
}

// utils, to move into a separate package?

func sqrt(x float64) float64 {
	if x == 0 || x == 1 {
		return x
	}

	// Start with an initial guess
	guess := x / 2.0
	prevGuess := 0.0

	// Use a small threshold to determine when to stop the approximation
	const threshold = 0.00001

	for abs(guess-prevGuess) > threshold {
		prevGuess = guess
		guess = 0.5 * (guess + x/guess)
	}

	return guess
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
