package gnofundme

import "std"

type Campaign struct{
	Title: string
	Description: string
	Goal: uint64
	Current: uint64
	Deadline: time
	Owner: *User
	Beneficiary: *User
	Contributors: *avl.Tree
}

type User struct{
	Name: string
	Email: string
	Address: std.Address
	Campaigns: []Campaign
}

type Contributor struct {
	User: *User
	Date: time
	Amount: uint64
	Display: bool
}

func NewUser(name string, email string, address std.Address) *User{
	return &User{
		Name: name,
		Email: email,
		Address: address,
	}
}

func NewCampaign(title string, description string, goal uint64, deadline time, owner *User, beneficiary *User) *Campaign{
	return &Campaign{
		Title: title,
		Description: description,
		Goal: goal,
		Current: 0,
		Deadline: deadline,
		Owner: owner,
		Beneficiary: beneficiary,
		Contributors: avl.NewTree(),
	}
}

func AddContributor(campaign *Campaign, user *User, amount uint64, display bool){
	campaign.Current += amount
	key := campaign.Contributors.Size()
	campaign.Contributors.Set(strconv.Itoa() &Contributor{
		User: user,
		Date: time.Now(),
		Amount: amount,
		Display: display,
	})
}