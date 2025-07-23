package helloworld

import "strings"

func Render(path string) string {
	content := `
## Example 1: HelloWorld

Let's start with a simple example to understand how to manually transpile a Solidity smart contract into Gno. This example defines a contract that stores a message and allows users to update it.

It's simple, but with this tutorial, youll already get in touch with these 3 important ideas:
1. State persistence
2. Public visibility
3. String handling

---

### In Solidity

^^^solidity
contract HelloWorld {
    string public message = "Hello, World!";

    function setMessage(string memory newMessage) public {
        message = newMessage;
    }
}
^^^

This Solidity contract does two things:

* It declares a state variable \^message\^ of type \^string\^, with a default value.
* Because it's marked \^public\^, Solidity automatically generates a getter function, so anyone can query its value.
* It also defines a \^setMessage\^ function, which updates the message using a string provided by the user. The \^memory\^ keyword means the string is passed temporarily and not stored between calls.

This is standard Solidity: public state variables, automatic accessors, and explicit memory management.

---

### Gno Version

^^^go
package helloworld

var message string = "Hello, World!"

func SetMessage(_ realm, newMessage string) {
    message = newMessage
}
^^^

This Gno version achieves the same behavior, but with different design principles:

* The variable ^message^ is declared at the top level. In Gno, top-level variables are **persistent**, meaning their values are stored across transactions, just like state variables in Solidity.
* The function ^SetMessage^ takes a string argument and reassigns ^message^. This function can be called from outside the package (i.e., by transactions on the blockchain).
* There's no automatic getter for ^message^. If you want users to be able to read the message from outside, you have to **manually write a ^GetMessage()^ function** like so:

^^^go
func GetMessage(_ realm) string {
    return message
}
^^^


### Key Concepts and Takeaways

1. **Public Variables vs Explicit Accessors** -
   In Solidity, adding ^public^ to a variable automatically generates a getter. In Gno, there's no such shortcut: you must define your own getter function if needed.

2. **Persistent State** -
   Both Solidity and Gno support persistent state, but they approach it differently. In Solidity, variables declared inside a contract are part of the blockchain storage. In Gno, top-level variables in a package act the same way: they live on-chain and retain values between function calls.

3. **No memory/storage distinction** -
   Solidity has different storage classes (^memory^, ^storage^, ^calldata^) that affect how variables are handled in memory. Gno, on the other hand, uses Go-native types and memory management. This simplifies things: if you're familiar with Go, you don't need to learn new storage rules.

4. **Go-inspired syntax** -
   Gno is syntactically close to Go, but you should always remember that it's not just "Go on-chain": it's a blockchain-specific dialect with deterministic execution and strict safety rules.

---

This HelloWorld contract is a perfect starting point. As we explore more complex examples, we'll introduce additional concepts like mappings, modifiers, and events: and see how to express each of them idiomatically in Gno.

`
	return strings.ReplaceAll(content, "^", "`")
}