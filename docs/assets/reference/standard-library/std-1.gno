// returns the list of coins owned by the address
GetCoins(addr Address) (dst Coins)

// sends coins from one address to another
SendCoins(from, to Address, amt Coins)

// returns the total supply of the coin
TotalCoin(denom string) int64

// issues coins to the address
IssueCoin(addr Address, denom string, amount int64)

// burns coins from the address
RemoveCoin(addr Address, denom string, amount int64)
