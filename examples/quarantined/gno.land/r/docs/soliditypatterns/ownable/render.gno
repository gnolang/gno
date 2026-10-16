package ownable

import "strings"

func Render(path string) string {
	content := `
# Ownable

This Solidity contract defines a counter that can be incremented or decremented by anyone.

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
package onlyownercanchange

var owner address = ""

func InitOwner(cur realm) {
	if len(owner) != 0 {
		panic("owner already defined")
	}
	owner = cur.Previous().Address()
}

func assertOnlyOwner(caller address) {
	if caller != owner {
		panic("caller isn't the owner")
	}
}

func ChangeOwner(cur realm, newOwner address) {
	assertOnlyOwner(cur.Previous().Address())
	if !newOwner.IsValid() {
		panic("invalid owner address")
	}
	owner = newOwner
}
^^^

* We declare a persistent global variable ^owner^ of type ^address^ (an alias for ^string^ in Gno), initialized to the empty string ^""^. This holds the address of the contract owner.
* The function ^InitOwner^ is used to initialize the ^owner^. It can only be called once: if ^owner^ is already set (non-empty), it throws an error using ^panic^. It sets the ^owner^ to the address of the previous realm (the caller) via ^cur.Previous().Address()^, which is similar to ^msg.sender^ in Solidity.
* The helper function ^assertOnlyOwner^ checks that the caller is the current ^owner^ by comparing the caller address it is given with ^owner^. If not, it panics. This is equivalent to a Solidity ^modifier onlyOwner^.
* The function ^ChangeOwner^ allows the current owner to transfer ownership to a new address. It first calls ^assertOnlyOwner(cur.Previous().Address())^ to ensure that only the current owner can do this, then refuses a ^newOwner^ that is not a valid address.
* All functions are capitalized (exported), meaning they can be called externally. However, ^assertOnlyOwner^ is unexported (lowercase), meaning it's only usable within the package.


> [!NOTE]
> For a library that handles this for you, check out [^/p/demo/ownable^](/p/demo/ownable).

---
`
	return strings.ReplaceAll(content+RenderDemo(), "^", "`")
}
