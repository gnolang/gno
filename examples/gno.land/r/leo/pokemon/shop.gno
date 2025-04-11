package pokemon

import (
	"std"
	"strconv"

	"gno.land/p/leo/utils"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/txlink"
)

var (
	itemData = map[ItemType]Item{
		SmallPotion: {300, "Small Potion", "Restores 20 HP for a Pokémon", applySmallPotion},
		BigPotion:   {700, "Big Potion", "Restores 60 HP for a Pokémon", applyBigPotion},
		MaxPotion:   {2500, "Max Potion", "Fully recovers the HP of a Pokémon", applyMaxPotion},
		Revive:      {1500, "Revive", "Revive a Pokémon and restores half its HP", applyRevive},
		MaxRevive:   {3500, "Max Revive", "Revive a Pokémon and fully recovers its HP", applyMaxRevive},
		Elixer:      {1000, "Elixer", "Restores 5 PP of each ability of one Pokémon", applyElixer},
		MaxElixer:   {2000, "Max Elixer", "Restores all the PP of each ability of one Pokémon", applyMaxElixer},
	}
)

func applySmallPotion(p *Pokemon) {
	if p.pv <= 0 {
		return
	}
	p.pv += 20
	maxPv := p.getDefault().pv
	if p.pv > maxPv {
		p.pv = maxPv
	}
}

func applyBigPotion(p *Pokemon) {
	if p.pv <= 0 {
		return
	}
	p.pv += 60
	maxPv := p.getDefault().pv
	if p.pv > maxPv {
		p.pv = maxPv
	}
}

func applyMaxPotion(p *Pokemon) {
	if p.pv <= 0 {
		return
	}
	p.pv = p.getDefault().pv
}

func applyRevive(p *Pokemon) {
	maxPv := p.getDefault().pv
	if p.pv < maxPv/2 {
		p.pv = maxPv / 2
	}
}

func applyMaxRevive(p *Pokemon) {
	maxPv := p.getDefault().pv
	p.pv = maxPv
}

func applyElixer(p *Pokemon) {
	defaultAttacks := p.getDefault().attacks
	for i := range p.attacks {
		p.attacks[i].pc += 5
		if p.attacks[i].pc > defaultAttacks[i].pc {
			p.attacks[i].pc = defaultAttacks[i].pc
		}
	}
}

func applyMaxElixer(p *Pokemon) {
	defaultAttacks := p.getDefault().attacks
	for i := range p.attacks {
		p.attacks[i].pc = defaultAttacks[i].pc
	}
}

func BuyItem(itemId ItemType) {
	usr := std.PreviousRealm().Address()
	pokemonDAO.AsserUserIsMember(usr)

	idx := utils.AddressSliceIndex(pokemonDAO.GetMembers(), func(_ int, member std.Address) bool {
		return usr == member
	})
	me := &pokemonDAO.userData[idx]
	if me.pokedollar < uint(itemData[itemId].price) {
		panic("not enough pokedollars")
	}
	me.inventory = append(pokemonDAO.userData[idx].inventory, itemId)
	me.pokedollar -= uint(itemData[itemId].price)
}

func renderPokedollar(price int) string {
	return md.Bold(strconv.Itoa(price) + " ₽")
}

func renderShop() (rend string) {
	table := mdtable.Table{Headers: []string{"Item", "Description", "Price"}}
	for k, v := range itemData {
		link := txlink.Call("BuyItem", "itemId", strconv.FormatInt(int64(k), 10))
		table.Append([]string{md.Link("Buy "+v.name, link), v.description, renderPokedollar(v.price)})
	}
	rend += table.String()
	return
}

func renderPokemonShop(params map[string]string) (rend string) {
	rend += md.H2("Shop")
	rend += md.H3("Me")
	idx := utils.AddressSliceIndex(pokemonDAO.GetMembers(), func(_ int, member std.Address) bool {
		return std.Address(params["user"]) == member
	})
	me := &pokemonDAO.userData[idx]
	rend += renderWallet(me.pokedollar)
	rend += renderInventory(me.inventory) + nl
	rend += renderShop() + nl
	rend += md.Link("Go back to lobby", atm_path+":lobby?user="+params["user"])
	return
}
