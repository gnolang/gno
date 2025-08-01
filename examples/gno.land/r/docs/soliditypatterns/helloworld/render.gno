package helloworld

import "strings"

func Render(path string) string {
	content := `
# HelloWorld

This example defines a contract that stores a message and allows users to update it.

---

## In Solidity

^^^solidity
contract HelloWorld {
    string public message = "Hello, World!";

    function setMessage(string memory newMessage) public {
        message = newMessage;
    }
}
^^^

* It declares a state variable ^message^ of type ^string^, with a default value, readle by anyone throug a getter thanks to public keyword
* It defines a ^setMessage^ function which updates the message using a string provided by the user. The ^memory^ keyword means the string is passed temporarily and not stored between calls.

---

## Gno Version

^^^go
package helloworld

var Message string = "Hello, World!"

func SetMessage(cur realm, newMessage string) {
    Message = newMessage
}
^^^

* We define a global variable ^Message^ of type ^int64^. Since it is declared at the package level, its value is **persistent**, just like in Solidity. As it is capitalized, it can be accessed from outside the package (read-only).
* The function ^SetMessage^ takes two arguments, the current ^realm^, passed automaticcaly at transaction, and a string ^newMessage^, and updates the ^Message^. As it is capitalized, it can be called from outside the package.
---
`
	return strings.ReplaceAll(content+RenderDemo(), "^", "`")
}
