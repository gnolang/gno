package statelock

import "strings"

func Render(path string) string {
	content := `
# State Lock

This example showcases how a Solidity modifier might be implemented in Gno.

---

## In Solidity

^^^solidity
contract StateLock {
    enum State { Active, Locked, Disabled }

    State public state = State.Active;

    modifier inState(State _state) {
        require(state == _state);
        _;
    }

    function lock() public inState(State.Active) {
        state = State.Locked;
    }

    function unlock() public inState(State.Locked) {
        state = State.Active;
    }
}
^^^

* An ^enum State^ is declared with three possible values: ^Active^, ^Locked^, and ^Disabled^.
* A ^public^ state variable ^state^ is declared to hold the current state. It can be read externally thanks to the ^public^ keyword.
* A ^modifier inState^ is defined to restrict function execution based on the current ^state^. It checks that ^state == _state^ using ^require^, and then allows the function body to continue.
* The ^lock^ function can only be called if the contract is in the ^Created^ state. When called, it changes the state to ^Locked^.


---

## Gno Version

^^^go
type State string

const (
	StateActive   State = "Active"
	StateLocked   State = "Locked"
	StateDisabled State = "Disabled"
)

var state State = StateActive

func assertInState(expected State) {
	if state != expected {
		panic("invalid state: expected " + expected + ", got " + state)
	}
}

func Lock(cur realm) {
	assertInState(StateActive)
	state = StateLocked
}

func Unlock(cur realm) {
	assertInState(StateLocked)
	state = StateActive
}
^^^

* We define a custom type ^State^ as a ^string^ to simulate enum-like behavior.
* We declare a ^const^ block listing all valid states: ^StateActive^, ^StateLocked^, and ^StateDisabled^. These are as the Solidity ^Enum^.
* A global variable ^state^ of type ^State^ is initialized to ^StateActive^.
* The helper function ^assertInState^ ensures that the contract is in the expected state. If not, it throws an error using ^panic^, showing both the expected and actual state values. This is equivalent to a Solidity ^modifier^.
* The ^Lock^ function changes the state from ^Active^ to ^Locked^. It first ensures the state is currently ^Active^ using ^assertInState^.
* The ^Unlock^ function reverses the state from ^Locked^ to ^Active^, again checking the current state before proceeding.

---

`
	return strings.ReplaceAll(content+RenderDemo(), "^", "`")
}
