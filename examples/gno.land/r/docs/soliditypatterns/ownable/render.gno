package ownable

import "strings"

func Render(path string) string {
	content := `
# Ownable

This Solidity contract defines a single-owner access-control pattern: the deployer becomes the owner, and only the owner may transfer ownership to a new address.

---

### Solidity Version

^^^solidity
contract OnlyOwnerCanChange {
    address private owner;

    constructor() {
        owner = msg.sender;
    }

    modifier onlyOwner {
        require(msg.sender == owner, "Not owner");
        _;
    }

    function changeOwner(address memory newOwner) public onlyOwner {
        owner = newOwner
    }
}
^^^

* It declares a state variable ^owner^ of type ^address^, which stores the address of the account that deployed the contract.
* The ^constructor^ sets the ^owner^ variable to the address that deployed the contract using ^msg.sender^.
* It defines a ^modifier onlyOwner^ that restricts function execution to the owner only, using ^require^ to check that the caller is the current ^owner^. The special symbol ^_;^ is where the modified function's code runs.
* It defines a ^changeOwner^ function that allows the current owner to assign a new owner address. The function uses the ^onlyOwner^ modifier to enforce access control.

---
## Gno Version

^^^go
package ownable

import (
	"chain"
)

var owner address

func InitOwner(cur realm) {
	if !cur.IsCurrent() {
		panic("spoofed realm")
	}
	if len(owner) != 0 {
		panic("owner already defined")
	}
	owner = cur.Previous().Address()
}

func ChangeOwner(cur realm, newOwner address) {
	if !cur.IsCurrent() {
		panic("spoofed realm")
	}
	if cur.Previous().Address() != owner {
		panic("caller isn't the owner")
	}
	owner = newOwner
	chain.Emit("OwnershipChange", "newOwner", newOwner.String())
}
^^^

* We declare a persistent global variable ^owner^ of type ^address^ (an alias for ^string^ in Gno). Realm globals are the persistence layer, so this value survives across transactions.
* Both ^InitOwner^ and ^ChangeOwner^ are **crossing functions** — their first parameter is ^cur realm^, a capability token for the live call into this realm. (You could instead claim ownership at deploy time inside ^init()^ via ^unsafe.OriginCaller()^, which is the closer analogue to a Solidity ^constructor^; here we use an explicit claim-once function so the flow is easy to exercise.)
* ^InitOwner^ lets the first caller claim ownership. It panics if ^owner^ is already set, so it works only once.
* ^ChangeOwner^ transfers ownership. Both functions authenticate the caller the same way, before touching state:
    * ^cur.IsCurrent()^ is the authentication primitive. It is **true only when ^cur^ matches the topmost live crossing frame**. A stale or forged realm value returns ^false^. Without this guard, an attacker could pass a stored realm value whose ^.Address()^ still resolves but no longer names the live caller.
    * ^cur.Previous().Address()^ then identifies the immediate caller — for ^ChangeOwner^ it is compared against ^owner^.
* On success ^ChangeOwner^ updates ^owner^ and emits an ^OwnershipChange^ event so off-chain indexers can observe the transfer.

> [!WARNING]
> **Do not translate ^msg.sender^ to a stack-walking helper.** A natural-looking port writes a non-crossing helper like ^func assertOnlyOwner() { if unsafe.PreviousRealm().Address() != owner { panic(...) } }^ and calls it from ^ChangeOwner^. This is **insecure**. A non-crossing call creates no realm boundary, so ^unsafe.PreviousRealm()^ returns whatever was previous *before the caller's own call* — not the immediate caller. An intermediate realm can launder the owner check. Caller identity must be derived inside the crossing function via ^cur.IsCurrent()^ + ^cur.Previous().Address()^, never in a non-crossing helper.

> [!NOTE]
> For a library that handles single-owner access control for you, check out [^gno.land/p/nt/ownable/v0^](/p/nt/ownable/v0).

---
`
	return strings.ReplaceAll(content+RenderDemo(), "^", "`")
}
