package emitevents

import "strings"

func Render(path string) string {
	content := `
# Emit Events

This Solidity contract defines a counter that emits an event when incremented.

---

### Solidity Version

^^^solidity
contract EmitEvents {
    uint public count;

    event Incremented(uint newValue);

    function increment() public {
        count++;
        emit Incremented(count);
    }
}
^^^
* ^event Incremented(uint newValue);^ → defines an event named ^Incremented^.
* ^emit Incremented(count);^ → triggers the event and stores the value in the blockchain logs.
* Events do not change the contract state, they are used to notify external observers.

---

### Gno Version

^^^go
package emitevents

import (
	"std"
	"strconv"
)

var Count int64 = 0

func Increment(cur realm) {
    Count++
    std.Emit("Incremented", "count", strconv.FormatInt(Count, 10))
}
^^^

* ^Count^ → persistent state variable.
* ^std.Emit("Incremented", "count", strconv.FormatInt(Count, 10))^
	* First argument ^"Incremented"^ is the event name.
	* ^"count"^ is the key name for the data.
	* ^strconv.FormatInt(Count, 10)^ converts the integer count to a string (required because ^std.Emit^ expects strings).
* No explicit event declaration needed in Gno.
* Any client can listen for this event by filtering ^"Incremented"^.

---
`
	return strings.ReplaceAll(content+RenderDemo(), "^", "`")
}
