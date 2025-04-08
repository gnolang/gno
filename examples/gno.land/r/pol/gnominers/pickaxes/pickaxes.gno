package pickaxes

type IPickaxe interface {
	HowMuchMining() uint
	String() string
	Clone() IPickaxe
}

// normal pickaxe

type NormalPickaxe struct {
	IPickaxe
}

func (p *NormalPickaxe) HowMuchMining() uint {
	return 1
}

func (p *NormalPickaxe) String() string {
	return "pickaxe"
}

func (p NormalPickaxe) Clone() IPickaxe {
	return &NormalPickaxe{}
}

// 3

type TriPickaxe struct {
	IPickaxe
}

func (p *TriPickaxe) HowMuchMining() uint {
	return 3
}

func (p *TriPickaxe) String() string {
	return "triPickaxe"
}

func (p TriPickaxe) Clone() IPickaxe {
	return &TriPickaxe{}
}

// 5

type PentaPickaxe struct {
	IPickaxe
}

func (p *PentaPickaxe) HowMuchMining() uint {
	return 5
}

func (p *PentaPickaxe) String() string {
	return "pentaPickaxe"
}

func (p PentaPickaxe) Clone() IPickaxe {
	return &PentaPickaxe{}
}

// 10

type DecaPickaxe struct {
	IPickaxe
}

func (p *DecaPickaxe) HowMuchMining() uint {
	return 10
}

func (p *DecaPickaxe) String() string {
	return "decaPickaxe"
}

func (p DecaPickaxe) Clone() IPickaxe {
	return &DecaPickaxe{}
}

// what the hell ?

type BirthdayPickaxe struct {
	IPickaxe
}

func (p *BirthdayPickaxe) HowMuchMining() uint {
	return 2005
}

func (p *BirthdayPickaxe) String() string {
	return "birthdayPickaxe"
}

func (p BirthdayPickaxe) Clone() IPickaxe {
	return &BirthdayPickaxe{}
}
