package fomo3d

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/ufmt"

	"gno.land/r/sys/users"
)

// RenderHome renders the main game state
func RenderHome() string {
	var builder strings.Builder
	builder.WriteString("# FOMO3D - The Ultimate Game of Greed\n\n")

	// About section
	builder.WriteString("## About the Game\n\n")
	builder.WriteString("FOMO3D is a game that combines elements of lottery and investment mechanics. ")
	builder.WriteString("Players purchase keys using GNOT tokens, where each key purchase:\n\n")
	builder.WriteString("* Extends the game timer\n")
	builder.WriteString("* Increases the key price by 1%\n")
	builder.WriteString("* Makes you the potential winner of the jackpot\n")
	builder.WriteString("* Distributes dividends to all key holders\n\n")
	builder.WriteString("## How to Win\n\n")
	builder.WriteString("* Be the last person to buy a key before the timer expires!\n\n")
	builder.WriteString("**Rewards Distribution:**\n")
	builder.WriteString("* 47% goes to the jackpot (for the winner)\n")
	builder.WriteString("* 28% distributed as dividends to all key holders\n")
	builder.WriteString("* 20% goes to next round's starting pot\n")
	builder.WriteString("* 5% development fee for continuous improvement\n\n")

	// Play Game section
	builder.WriteString("## How to Play\n\n")
	builder.WriteString(ufmt.Sprintf("1. **Buy Keys** - Send GNOT to this realm with function [`BuyKeys()`](%s)\n", gameState.BuyKeysLink))
	builder.WriteString(ufmt.Sprintf("2. **Collect Dividends** - Call [`ClaimDividends()`](%s) to collect your earnings\n", gameState.ClaimDividendsLink))
	builder.WriteString("3. **Check Your Stats** - Append `:player/` followed by your address or namespace to the current URL to view your keys and dividends\n")
	if gameState.Ended {
		builder.WriteString(ufmt.Sprintf("4. **Start New Round** - Call [`StartGame()`](%s) to begin a new round\n", gameState.StartGameLink))
	}
	builder.WriteString("\n")

	// Game Status section
	builder.WriteString("## Game Status\n\n")
	if gameState.StartBlock == 0 {
		builder.WriteString("🔴 Game has not started yet.\n\n")
	} else {
		if gameState.Ended {
			builder.WriteString("🔴 **Game Status:** Ended\n")
			builder.WriteString(ufmt.Sprintf("🏆 **Winner:** %s\n\n", gameState.LastBuyer))
		} else {
			builder.WriteString("🟢 **Game Status:** Active\n\n")
			builder.WriteString(ufmt.Sprintf("🔄 **Round:** %d\n\n", gameState.CurrentRound))
			builder.WriteString(ufmt.Sprintf("⏱️ **Time Remaining:** %d blocks\n\n", gameState.EndBlock-std.ChainHeight()))
		}
		builder.WriteString(ufmt.Sprintf("💰 **Jackpot:** %d ugnot\n\n", gameState.Jackpot))
		builder.WriteString(ufmt.Sprintf("🔑 **Key Price:** %d ugnot\n\n", gameState.KeyPrice))
		builder.WriteString(ufmt.Sprintf("📊 **Total Keys:** %d\n\n", gameState.TotalKeys))
		builder.WriteString(ufmt.Sprintf("👤 **Last Buyer:** %s\n\n", getDisplayName(gameState.LastBuyer)))
		builder.WriteString(ufmt.Sprintf("🎮 **Next Round Pot:** %d ugnot\n\n", gameState.NextPot))
	}

	// Separator before less important sections
	builder.WriteString("---\n\n")

	// Vote For Me section
	builder.WriteString("### Vote For Us! 🗳️\n\n")
	builder.WriteString("If you enjoy playing FOMO3D, please consider upvoting this game in the [Hall of Realms](https://gno.land/r/leon/hor)!\n\n")
	builder.WriteString("Your support helps more players discover the game and grow our community! 🚀\n\n")

	// Report Bug section
	builder.WriteString("### Report a Bug 🪲\n\n")
	builder.WriteString("Something unusual happened? Help us improve the game by reporting bugs!\n")
	builder.WriteString("[Visit our GitHub repository](https://github.com/gnolang/gno/issues)\n\n")
	builder.WriteString("Please include:\n")
	builder.WriteString("* Detailed description of what happened\n")
	builder.WriteString("* Transaction hash (if applicable)\n")
	builder.WriteString("* Your address\n")
	builder.WriteString("* Current round number\n")

	return builder.String()
}

// RenderPlayer renders specific player information
func RenderPlayer(addr address, keys int64, dividends int64) string {
	var builder strings.Builder
	displayName := getDisplayName(addr)
	builder.WriteString(ufmt.Sprintf("# Player Stats: %s\n\n", displayName))
	builder.WriteString("## Your Holdings\n\n")
	builder.WriteString(ufmt.Sprintf("🔑 **Keys Owned:** %d\n\n", keys))
	builder.WriteString(ufmt.Sprintf("💰 **Unclaimed Dividends:** %d ugnot\n\n", dividends))

	// Check if player has any NFTs
	nftBalance, err := BalanceOf(addr)
	if err == nil && nftBalance > 0 {
		builder.WriteString("## Your Victory NFTs 🏆\n\n")

		// Iterate through all rounds up to current round to find player's NFTs
		for i := int64(1); i <= gameState.CurrentRound; i++ {
			tokenID := grc721.TokenID(strconv.FormatInt(i, 10))
			owner, err := OwnerOf(tokenID)
			if err == nil && owner == addr {
				metadata, err := TokenMetadata(tokenID)
				if err == nil {
					builder.WriteString(ufmt.Sprintf("### Round #%d Winner\n", i))
					builder.WriteString(ufmt.Sprintf("![NFT](%s)\n\n", metadata.Image))
					builder.WriteString("---\n\n")
				}
			}
		}
	}

	builder.WriteString("## Actions\n\n")
	builder.WriteString(ufmt.Sprintf("* To buy more keys, send GNOT to this realm with [`BuyKeys()`](%s)\n", gameState.BuyKeysLink))
	if dividends > 0 {
		builder.WriteString("* You have unclaimed dividends! Call `ClaimDividends()` to collect them\n")
	}

	return builder.String()
}

// Helper to get display name - just returns namespace if exists, otherwise address
func getDisplayName(addr address) string {
	if user := users.ResolveAddress(addr); user != nil {
		return user.Name()
	}
	return addr.String()
}

// UpdateFunctionLinks updates the links for game functions
func UpdateFunctionLinks(buyKeysLink string, claimDividendsLink string, startGameLink string) {
	Ownable.AssertOwnedByPrevious()
	gameState.BuyKeysLink = buyKeysLink
	gameState.ClaimDividendsLink = claimDividendsLink
	gameState.StartGameLink = startGameLink
}
