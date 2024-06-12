package flippando

import (
	"time"
	"strconv"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/flippandoserver"
)

// realm state
var (	
	gameStore     avl.Tree // string (game ID) -> *Game
	gameIDCounter uint64

	user2Games avl.Tree // std.Address -> []*Game
	marketplace = NewMarketplace()
)

const zeroes = "000000000"

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
	
	//games := GetUserGames(player)
	// Set up Game struct. Save in gameStore and user2games.
	gameIDCounter++
	board, gameTiles, boardSize := flippandoserver.CreateGameBoard(tileType, boardSize, time.Now().Unix())
	
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
	var currentGame *Game // Declare a pointer to a Game
	for _, game := range games {
		if game.ID == gameID {
			currentGame = game
			break // Stop the loop once the game is found
		}
	}
	//currentGame := games[0]
	currentGamePositions, err := ParseString(positions)
	
	// call real FlipTiles on flippandoserver, which does the game mechanics
	// and returns 2 []int64 boards: the GameBoard, and the SolvedGameBoard
	
	board, solvedBoard := flippandoserver.FlipTiles(currentGame.GameBoard, 
		currentGame.SolvedGameBoard, 
		currentGame.GameTiles, 
		currentGamePositions,
		time.Now().Unix())
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

func CreateBasicNFT(player std.Address, gameId string) string {
	// call CreateBasicFlipNFT from basicNFT.gno
	// call MintFlipFungibleToken from token,gno
	// lock the Flip fungible token inside the basicNFT in nft2Flip avl.tree
	metadata, errString := CreateBasicFlipNFT(player, gameId)
	if errString != "" {
		return errString
	}
	return metadata
}

func GetUserBasicNFTs(addr std.Address, readyToUse string) string {
	// call GetUserBasicFlipNFTs from basicNFT.gno
	userNFTs := GetUserBasicFlipNFTs(addr, readyToUse)
	return userNFTs
}

// if excludePlayer is != "" return all nfts except those of excludePlayer
// used in displaying available nfts for art
// if excludePlayer == "" return all nfts, used for stats, etc
func GetAllBasicNFTs(excludePlayer std.Address) string {
	allNFTs := GetAllBasicFlipNFTs(excludePlayer)
	return allNFTs
}

func CreateCompositeNFT(player std.Address, cWidth string, cHeight string, bTokenIDsArray string) string {
	metadata, errString := CreateCompositeFlipNFT(player, cWidth, cHeight, bTokenIDsArray)
	
	if errString != "" {
		return errString
	}
	return metadata
}
    
func GetUserGames(user std.Address) []*Game {
	val, exist := user2Games.Get(user.String())
	//var games []*Game
	if !exist {
		return nil
	}
	return val.([]*Game)
}

func GetUserGamesByStatus(user std.Address, status string) string {
	val, exist := user2Games.Get(user.String())
	if !exist {
		return "{\"userGames\":[]}"
	}

	allGames := val.([]*Game)
	var matchingGames []Game
	for _, game := range allGames {
		if game.GameStatus == status {
			matchingGames = append(matchingGames, *game) 
		}
	}

	ugr := UserGamesResult{
		UserGames: matchingGames,
	}

	ugrJSON, err := ugr.MarshalJSON()
	if err != nil {
		panic(err)
	}

	return string(ugrJSON)
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

func GetArtListings() string {
	return marketplace.GetListings()
}

func ListMarketplaceNFT(compositeTokenID string, seller std.Address, price string) error {
	return marketplace.ListNFT(grc721.TokenID(compositeTokenID), seller, price);
}

func BuyMarketplaceNFT(buyer std.Address, compositeTokenId string, ) (buyNFTResult string) {
	return marketplace.BuyNFT(buyer, grc721.TokenID(compositeTokenId)) 
}

func RemoveMarketplaceNFT(compositeTokenID string, seller std.Address) error {
	return marketplace.RemoveListing(grc721.TokenID(compositeTokenID), seller)
}

func ResetGame(gameId string) {
	//- utility function, nullifies the game[gameId]
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


