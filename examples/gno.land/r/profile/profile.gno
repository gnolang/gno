package profile

import (
	"std"

	"gno.land/r/users"
)

func newProfile(addr std.Address) *Profile {
	now := std.GetTime()
	return &Profile{
		address: addr,
		created: now,
	}
}

type Profile struct {
	dict    map[string]interface{}
	address std.Address
	created uint64
	updated uint64
}

func (p *Profile) save() {
	now := std.GetTime()
	if p.created == 0 {
		p.created = now
	}
	p.updated = now
}

func (p *Profile) update(dict map[string]interface{}) {
	for k, v := range dict {
		// TODO: additional checks here
		if k == "address" || k == "created" || k == "updated" {
			panic("reserved profile key")
		}
		p.dict[k] = v
	}
	p.save()
}

func (p *Profile) Render() string {
	output := ""
	output += ufmt.Sprintf("* address: %q\n", p.address)
	output += ufmt.Sprintf("* created: %v\n", p.created)
	if p.updated > 0 {
		output += ufmt.Sprintf("* updated: %v\n", p.updated)
	}
	for k, v := range p.dict {
		output += ufmt.Sprintf("* %s: %v\n", k, v)
	}
	return output
}
