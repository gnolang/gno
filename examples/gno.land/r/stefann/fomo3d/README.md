# FOMO3D Game

FOMO3D (Fear Of Missing Out 3D) is a blockchain-based game that combines elements of a lottery and investment mechanics. Players purchase keys using GNOT tokens, where each key purchase:

- Extends the game timer
- Increases the key price by 1%
- Makes the buyer the potential winner of the jackpot
- Distributes dividends to all key holders

## Game Mechanics

- The last person to buy a key before the timer expires wins the jackpot (47% of all purchases)
- Key holders earn dividends from each purchase (28% of all purchases)
- 20% of purchases go to the next round's starting pot
- 5% goes to development fee
- Game ends when the timer expires

## How to Play

1. **Buy Keys** - Send GNOT to this realm with `BuyKeys` to purchase keys
2. **Collect Dividends** - Call `ClaimDividends` to collect your earnings
3. **Check Your Stats** - Append `:player/` followed by your address or namespace to the current URL to view your keys and dividends
4. **Start New Round** - Call `StartGame` to begin a new round (only available when game has ended)

## Game Parameters

- Minimum Key Price: 100,000 ugnot
- Time Extension: 86,400 blocks (~24 hours @ 1s blocks)
- Distribution:
  - Jackpot: 47%
  - Dividends: 28%
  - Next Round Pot: 20%
  - Development Fee: 5%
