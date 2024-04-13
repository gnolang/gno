package gnofundme

import "std"

type User struct {
	name      string
	email     string
	address   std.Address
	campaigns []*Campaign
}

func NewUser(name string, email string, address std.Address) *User {
	return &User{
		name:    name,
		email:   email,
		address: address,
	}
}

func (u User) GetName() string {
	return u.name
}

func (u User) GetEmail() string {
	return u.email
}

func (u User) GetAddress() std.Address {
	return u.address
}
