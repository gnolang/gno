package address

import (
	"fmt"
	"sort"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

type Entry struct {
	crypto.Address
	Names []string
}

type Book struct {
	addrsToNames map[crypto.Address][]string // address -> []names
	namesToAddrs map[string]crypto.Address   // name -> address
}

func NewBook() *Book {
	return &Book{
		addrsToNames: map[crypto.Address][]string{},
		namesToAddrs: map[string]crypto.Address{},
	}
}

func remove(s []string, i int) []string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func (bk *Book) Add(addr crypto.Address, name string) {
	// Check and register address if it wasn't existing
	names, ok := bk.addrsToNames[addr]
	if !ok {
		bk.addrsToNames[addr] = []string{}
	}

	// If name is empty, stop here
	if name == "" {
		return
	}

	oldAddr, ok := bk.namesToAddrs[name]
	if !ok {
		bk.namesToAddrs[name] = addr
		bk.addrsToNames[addr] = append(names, name)
		return
	}

	// Check if the association already exist
	if oldAddr.Compare(addr) == 0 {
		return // nothing to do
	}

	// If the name is associated with a different address, remove the old association
	oldNames := bk.addrsToNames[oldAddr]
	for i, oldName := range oldNames {
		if oldName == name {
			bk.addrsToNames[oldAddr] = remove(oldNames, i)
			break
		}
	}

	// Add the new association
	bk.namesToAddrs[name] = addr
	bk.addrsToNames[addr] = append(names, name)
}

func (bk Book) List() []Entry {
	entries := make([]Entry, 0, len(bk.addrsToNames))
	for addr, names := range bk.addrsToNames {
		namesCopy := make([]string, len(names))
		copy(namesCopy, names)

		newEntry := Entry{
			Address: addr,
			Names:   namesCopy,
		}

		// Find the correct place to insert newEntry using binary search.
		i := sort.Search(len(entries), func(i int) bool {
			return entries[i].Address.Compare(newEntry.Address) >= 0
		})

		entries = append(entries[:i], append([]Entry{newEntry}, entries[i:]...)...)
	}

	return entries
}

func (bk Book) GetByAddress(addr crypto.Address) (names []string, ok bool) {
	names, ok = bk.addrsToNames[addr]
	return
}

func (bk Book) GetByName(name string) (addr crypto.Address, ok bool) {
	addr, ok = bk.namesToAddrs[name]
	return
}

func (bk Book) GetFromNameOrAddress(addrOrName string) (addr crypto.Address, names []string, ok bool) {
	var err error
	if addr, ok = bk.namesToAddrs[addrOrName]; ok {
		names = []string{addrOrName}
	} else if addr, err = crypto.AddressFromBech32(addrOrName); err == nil {
		// addr is valid, now check if we have it
		names, ok = bk.addrsToNames[addr]
	}

	return
}

func (bk Book) ImportKeybase(path string) error {
	kb, err := keys.NewKeyBaseFromDir(path)
	if err != nil {
		return fmt.Errorf("unable to load keybase: %w", err)
	}
	defer kb.CloseDB()

	keys, err := kb.List()
	if err != nil {
		return fmt.Errorf("unable to list keys: %w", err)
	}

	for _, key := range keys {
		name := key.GetName()
		bk.Add(key.GetAddress(), name)
	}

	return nil
}
