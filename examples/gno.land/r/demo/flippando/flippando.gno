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

// creates game object and initializes board
func StartGame(player std.Address, tileType string, boardSize int) string {
	
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
	currentGamePositions, err := ParseString(positions)
	
	// calls real FlipTiles on flippandoserver, which does the game mechanics
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
	// calls CreateBasicFlipNFT from basicNFT.gno
	// calls MintFlipFungibleToken from token,gno
	// locks the Flip fungible token inside the basicNFT in nft2Flip avl.tree

	metadata, errString := CreateBasicFlipNFT(player, gameId)
	if errString != "" {
		return errString
	}
	return metadata
}

func GetUserBasicNFTs(addr std.Address, readyToUse string) string {
	// calls GetUserBasicFlipNFTs from basicNFT.gno
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

func GetGameByIDForUser(userAddress std.Address, gameID string) (*Game, error) {
	// Retrieve the list of games for the specified user
	gamesInterface, exists := user2Games.Get(userAddress.String())
	if !exists {
		return nil, ErrUserHasNoGames
	}

	// Type assertion to convert from interface{} to []*Game
	games, ok := gamesInterface.([]*Game)
	if !ok {
		return nil, ErrInvalidGameType
	}

	// Iterate through the slice of games to find the one with the matching ID
	for _, game := range games {
		if game.ID == gameID {
			return game, nil
		}
	}

	// Return error if the game is not found
	return nil, ErrGameWithIdNotFound
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

// marketplace wrapper

func GetArtListings() string {
	return marketplace.GetCompositeNFTListings()
}

func GetBasicListings() string {
	return marketplace.GetBasicNFTListings()
}

func ListCompositeNFT(compositeTokenID string, seller std.Address, price string) error {
	return marketplace.ListCompositeNFT(grc721.TokenID(compositeTokenID), seller, price);
}

func ListBasicNFT(basicTokenID string, seller std.Address, price string) error {
	return marketplace.ListBasicNFT(grc721.TokenID(basicTokenID), seller, price);
}

func BuyCompositeNFT(buyer std.Address, compositeTokenId string, ) (buyNFTResult string) {
	return marketplace.BuyCompositeNFT(buyer, grc721.TokenID(compositeTokenId)) 
}

func BuyBasicNFT(buyer std.Address, basicTokenId string, ) (buyNFTResult string) {
	return marketplace.BuyBasicNFT(buyer, grc721.TokenID(basicTokenId)) 
}

func RemoveCompositeNFTListing(compositeTokenID string, seller std.Address) error {
	return marketplace.RemoveCompositeNFTListing(grc721.TokenID(compositeTokenID), seller)
}

func RemoveBasicNFTListing(compositeTokenID string, seller std.Address) error {
	return marketplace.RemoveBasicNFTListing(grc721.TokenID(compositeTokenID), seller)
}

func IsBasicTokenIdListed(tokenId string) error {
	return marketplace.IsBasicTokenIdListed(tokenId)
}

func RenderHome() string {
	return "render path"
}


