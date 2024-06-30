package profile

import (
	"std"

	"gno.land/p/ufmt"
	// "gno.land/r/users"
)

// FIXME: manage privacy?

func newProfile(addr std.Address) *Profile {
	now := std.GetTimestamp()
	return &Profile{
		address: addr,
		created: now,
	}
}

type Profile struct {
	dict    map[string]interface{}
	address std.Address
	created std.Time
	updated std.Time
}

func (p *Profile) save() {
	now := std.GetTimestamp()
	if p.created == 0 {
		p.created = now
	}
	p.updated = now
}

func (p *Profile) update(dict map[string]interface{}) {
	for k, v := range dict {
		// TODO: additional checks here
		// TODO: check v.Type()
		// TODO: check v.Len()
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
