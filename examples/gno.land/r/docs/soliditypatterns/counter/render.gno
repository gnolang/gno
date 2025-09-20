package counter

import "strings"

func Render(path string) string {
	content := `
# Counter

This Solidity contract defines a counter that can be incremented or decremented by anyone.

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

This Solidity contract does:

* Declare a ^count^ variable readable by anyone initialized at 0
* Define two functions, ^increment^ and ^decrement^, that increase and decrease the counter by 1.
* These functions are marked ^public^, meaning anyone can call them

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

* We define a global variable ^count^ of type ^int64^. Since it is declared at the package level, its value is **persistent**, just like in Solidity. As it is capitalized, it can be accessed from outside the package (read-only).
* We provide two functions: ^Increment^ and ^Decrement^, which modify the counter value using Go-style syntax (^count++^ and ^count--^). As they are capitalized, they can be called from outside the package.
 
---
`
	return strings.ReplaceAll(content+RenderDemo(), "^", "`")
}
