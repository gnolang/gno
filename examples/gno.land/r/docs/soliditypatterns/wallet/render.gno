package wallet

import "strings"

func Render(path string) string {
	content := `
# Wallet

This example defines a contract that receive and send coins.

---

## In Solidity

^^^solidity
contract Wallet {
	receive() external payable {}

	function getBalance() public view returns (uint256) {
	    return address(this).balance;
	}

	function send(address payable _to) public payable {
        (bool sent, ) = _to.call{value: msg.value}("");
        require(sent, "Failed to send Ether");
    }
}
^^^

* It defines a ^receive^ function marked ^external payable^ with no body. This is a special function in Solidity that allows the contract to receive ETH sent without data.
* It defines a ^getBalance^ function marked ^public view^ that returns the ETH balance of the contract itself using ^address(this).balance^.
* It defines a ^send^ function that takes an address and an amount as parameters. The ^_to^ parameter is marked ^payable^ so it can receive ETH. It uses the low-level ^.call{value: amount}("")^ to send the specified ETH. The call returns a boolean indicating success and optional returned data. The function requires the send to succeed, otherwise it reverts with an error message.

---

## Gno Version

^^^go
package wallet

import "std"

func Receive(cur realm) {
	std.AssertOriginCall()
}

func GetBalance(cur realm) std.Coins {
	banker := std.NewBanker(std.BankerTypeReadonly)
	return banker.GetCoins(std.CurrentRealm().Address())
}

func Send(cur realm, to std.Address) {
	std.AssertOriginCall()

	banker := std.NewBanker(std.BankerTypeRealmSend)
	banker.SendCoins(
		std.CurrentRealm().Address(),
		to,
		std.OriginSend(),
	)
}
^^^

* It defines a ^Receive^ function that is called when the realm receives tokens. The ^std.AssertOriginCall()^ ensures the call originates from a valid transaction source rather than another smart contract or unauthorized origin.
* It defines a ^getBalance^ function that returns the current realm’s balance. It creates a banker object in read-only mode with ^std.NewBanker(std.BankerTypeReadonly)^ and queries the coins held by the realm’s address using ^GetCoins^.
* It defines a ^Send^ function that transfers coins from the realm to another address. It first verifies the call’s origin with ^std.AssertOriginCall()^, then creates a banker object with ^std.BankerTypeRealmSend^ to enable sending. It calls ^SendCoins^ to move tokens from the current realm’s address ^to^ the target address to, using the exact amount of coins provided by ^std.OriginSend()^.
---
`
	return strings.ReplaceAll(content+RenderDemo(), "^", "`")
}
