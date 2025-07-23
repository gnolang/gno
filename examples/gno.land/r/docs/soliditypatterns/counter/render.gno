package counter

import "strings"

func Render(path string) string {
	content := `
## Example 2: Counter

In this second example, we will see how to handle a numerical state and basic arithmetic operations. The original Solidity contract defines a counter that can be incremented or decremented by anyone. This type of contract is often used as a learning example for understanding state updates on the blockchain.

---

### Solidity Version

^^^solidity
contract Counter {
    int public count = 0;

    function increment() public {
        count++;
    }

    function decrement() public {
        count--;
    }
}
^^^

#### Explanation

Here's what this Solidity contract does:

* It defines a signed integer state variable ^count^, initialized to ^0^.
* The ^public^ keyword means a getter is automatically created, so external callers can query the current value of ^count^.
* It defines two functions, ^increment^ and ^decrement^, that respectively increase and decrease the counter by 1.
* These functions are marked ^public^, meaning any account can call them.

This is a very typical example of a mutable state variable being modified through transaction calls.

---

### Gno Version

^^^go
package counter

var count int64 = 0

func Increment(cur realm) {
    count++
}

func Decrement(cur realm) {
    count--
}
^^^

#### Explanation

Here's how we express the same logic in Gno:

* We define a top-level variable ^count^ of type ^int64^. Since it is declared at the package level, its value is **persistent**, just like in Solidity.
* We provide two functions: ^Increment^ and ^Decrement^, which modify the counter value using Go-style syntax (^count++^ and ^count--^).
* These functions are automatically callable from outside the module: there's no need to use a ^public^ keyword in Gno.
* Just like in the HelloWorld example, **if we want to allow external access to read the current value**, we must add a manual getter:

^^^go
func GetCount(cur realm) int64 {
    return count
}
^^^

---

### Key Concepts and Takeaways

1. **Gno uses Go-style types**
   Solidity's ^int^ is a 256-bit signed integer by default (^int256^). In Gno, you must choose an explicit Go type: for example, ^int64^ or ^int32^, depending on your needs. Gno does not have arbitrarily large integers like Solidity's ^int256^.

2. **State is persistent by default at the package level**
   When a variable is declared outside of any function in a Gno package, it becomes persistent across transactions: similar to a contract's state variable in Solidity.

3. **No visibility modifiers are needed**
   Solidity requires functions to be marked as ^public^, ^private^, or ^internal^. In Gno, all top-level functions are **public by default**: meaning they can be called by users through transactions or by other Gno packages.

4. **Manual accessors for state variables**
   Solidity's ^public^ variables automatically generate getter functions. In Gno, you must manually write getter functions to expose state. This gives developers more explicit control over what is visible and what is not.

5. **Simple and deterministic behavior**
   Since Gno builds on Go's syntax and semantics, there's no need for new keywords like ^memory^ or ^storage^. Gno programs are deterministic and written with simplicity in mind: if you're familiar with Go, this style will feel natural.

---

This counter example adds another important concept to our toolbox: arithmetic operations and type handling. As we continue, we'll explore more complex contract patterns like access control, mappings, and enums: and how Gno handles them with its minimalistic, Go-like approach.

`
	return strings.ReplaceAll(content+RenderDemo(), "^", "`")
}
