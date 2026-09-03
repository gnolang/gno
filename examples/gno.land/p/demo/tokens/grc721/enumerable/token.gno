package enumerable

import (
	"gno.land/p/demo/tokens/grc721"
)

const Kind = "enumerable"

// NewEnumerable attaches an enumerable extension and registers it as a core
// Extension. Attach it before the first mint so every token is indexed.
func NewEnumerable(coreLedger *grc721.PrivateLedger) (*Enumerable, *Ledger) {
	if coreLedger == nil {
		panic("enumerable: nil core ledger")
	}

	st := &storage{}
	core := coreLedger.ReadToken()
	led := &Ledger{core: core, st: st}
	view := &Enumerable{core: core, st: st}
	coreLedger.RegisterExtension(led, view)

	return view, led
}

func (e *Enumerable) ExtensionKind() string {
	return Kind
}

func (e *Enumerable) TokenID() string {
	return e.core.ID()
}

func (e *Enumerable) TotalSupply() int64 {
	return int64(len(e.st.allTokens))
}

func (e *Enumerable) TokenByIndex(index int64) (grc721.TokenID, error) {
	if index < 0 || index >= int64(len(e.st.allTokens)) {
		return "", ErrIndexOutOfRange
	}

	return e.st.allTokens[index], nil
}

func (e *Enumerable) TokenOfOwnerByIndex(owner address, index int64) (grc721.TokenID, error) {
	list := e.ownedList(owner)
	if list == nil || index < 0 || index >= int64(len(list.ids)) {
		return "", ErrIndexOutOfRange
	}

	return list.ids[index], nil
}

func (e *Enumerable) ownedList(owner address) *tokenList {
	v := e.st.owned.Get(owner.String())
	if v == nil {
		return nil
	}

	return v.(*tokenList)
}

func (led *Ledger) ExtensionKind() string { return Kind }

func (led *Ledger) OnMint(to address, tid grc721.TokenID) {
	led.addToAll(tid)
	led.addToOwner(to, tid)
}

func (led *Ledger) OnTransfer(from, to address, tid grc721.TokenID) {
	led.removeFromOwner(from, tid)
	led.addToOwner(to, tid)
}

func (led *Ledger) OnBurn(tid grc721.TokenID) {
	if v := led.st.owner.Get(tid.String()); v != nil {
		led.removeFromOwner(v.(address), tid)
	}

	led.removeFromAll(tid)
}

func (led *Ledger) addToAll(tid grc721.TokenID) {
	led.st.allIndex.Set(tid.String(), len(led.st.allTokens))
	led.st.allTokens = append(led.st.allTokens, tid)
}

// Swap-and-pop: move the last id into the freed slot, then truncate.
func (led *Ledger) removeFromAll(tid grc721.TokenID) {
	tidStr := tid.String()

	v := led.st.allIndex.Get(tidStr)
	if v == nil {
		return
	}

	idx := v.(int)

	last := len(led.st.allTokens) - 1
	if idx != last {
		moved := led.st.allTokens[last]
		led.st.allTokens[idx] = moved
		led.st.allIndex.Set(moved.String(), idx)
	}

	led.st.allTokens = led.st.allTokens[:last]

	led.st.allIndex.Remove(tidStr)
}

func (led *Ledger) addToOwner(owner address, tid grc721.TokenID) {
	ownerStr := owner.String()

	var list *tokenList
	if v := led.st.owned.Get(ownerStr); v != nil {
		list = v.(*tokenList)
	} else {
		list = &tokenList{}
		led.st.owned.Set(ownerStr, list)
	}

	led.st.ownedIndex.Set(tid.String(), len(list.ids))
	list.ids = append(list.ids, tid)
	led.st.owner.Set(tid.String(), owner)
}

// Swap-and-pop within the owner list; drops the owner entry when it empties.
func (led *Ledger) removeFromOwner(owner address, tid grc721.TokenID) {
	ownerStr := owner.String()

	v := led.st.owned.Get(ownerStr)
	if v == nil {
		return
	}

	list := v.(*tokenList)
	tidStr := tid.String()

	iv := led.st.ownedIndex.Get(tidStr)
	if iv == nil {
		return
	}

	idx := iv.(int)

	last := len(list.ids) - 1
	if idx != last {
		moved := list.ids[last]
		list.ids[idx] = moved
		led.st.ownedIndex.Set(moved.String(), idx)
	}

	list.ids = list.ids[:last]

	led.st.ownedIndex.Remove(tidStr)

	if len(list.ids) == 0 {
		led.st.owned.Remove(ownerStr)
	}

	led.st.owner.Remove(tidStr)
}
